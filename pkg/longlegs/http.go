package longlegs

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/rs/zerolog/log"
)

type ErrRedirect struct {
	Location string
}

func (err *ErrRedirect) Error() string {
	return fmt.Sprintf("Redirected %s", err.Location)
}

// decompressingTransport wraps an http.RoundTripper to automatically handle
// compressed responses (gzip, deflate, brotli)
type decompressingTransport struct {
	Transport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface
func (t *decompressingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Use default transport if none specified
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Add Accept-Encoding header to request
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}

	// Execute the request
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Skip decompression for HEAD requests (no body per HTTP spec)
	if req.Method == http.MethodHead {
		return resp, nil
	}

	// Wrap response body with appropriate decompressor
	encoding := resp.Header.Get("Content-Encoding")
	if encoding != "" {
		wrappedBody, err := wrapDecompressor(resp.Body, encoding)
		if err != nil {
			log.Warn().Msgf("Failed to create decompressor for %s: %v", encoding, err)
			return resp, nil // Return original response on decompression error
		}
		resp.Body = wrappedBody
		// Clear Content-Encoding since we've decompressed
		resp.Header.Del("Content-Encoding")
		// Content-Length is now unknown
		resp.ContentLength = -1
		resp.Header.Del("Content-Length")
	}

	return resp, nil
}

// wrapDecompressor wraps an io.ReadCloser with the appropriate decompression reader
func wrapDecompressor(body io.ReadCloser, encoding string) (io.ReadCloser, error) {
	// Handle multiple encodings (split by comma)
	encodings := strings.Split(encoding, ",")

	// Process encodings in reverse order (last applied first when decompressing)
	for i := len(encodings) - 1; i >= 0; i-- {
		enc := strings.TrimSpace(encodings[i])

		switch enc {
		case "br":
			// Brotli decompression
			return io.NopCloser(brotli.NewReader(body)), nil
		case "gzip":
			// Gzip decompression
			reader, err := gzip.NewReader(body)
			if err != nil {
				return nil, fmt.Errorf("failed to create gzip reader: %w", err)
			}
			return reader, nil
		case "deflate":
			// Deflate decompression
			return flate.NewReader(body), nil
		case "identity", "":
			// No compression
			continue
		default:
			// Unknown encoding - log warning but continue
			log.Warn().Msgf("Unknown Content-Encoding: %s", enc)
			continue
		}
	}

	// No recognized encoding, return original body
	return body, nil
}

// Call: `defer resp.Body.Close()`
func (spider *Spider) Get(url url.URL) (*http.Response, int64, error) {
	// TODO global last crawl check & wait here. Need to fix crawl.go
	// if !spider.CanCrawl(url) {
	// 	log.Error().Msg("Site blocked by robots.txt")
	// 	return nil, 0, errors.New("URL blocked by robots.txt")
	// }
	// spider.WaitCrawlDelay()

	client := &http.Client{
		Transport: &decompressingTransport{
			Transport: http.DefaultTransport,
		},
		Jar:     spider.cookieJar,
		Timeout: spider.requestTimeout,
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Info().Msgf("Failed to Request %s", url.String())
		return nil, 0, err
	}

	req.Header.Set("User-Agent", spider.userAgent)

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Info().Msgf("Failed to Request %s", url.String())
		return nil, 0, ErrRequestFailed
	}
	ms := time.Now().Sub(start).Milliseconds()
	return resp, ms, nil
}

// sameHostRedirect reports whether to and from are the same site. Hostname
// comparison ignores ports, so example.com:443 is not a different site from
// example.com.
func sameHostRedirect(from, to *url.URL) bool {
	if from == nil || to == nil {
		return false
	}
	fromHost := from.Hostname()
	toHost := to.Hostname()
	if fromHost == "" || toHost == "" {
		return false
	}
	return strings.EqualFold(fromHost, toHost)
}

// HeadNoRedirects issues a HEAD request and returns ErrRedirect instead of
// following redirects. A same-host canonical redirect is not logged: the
// Location may still echo a query string, and a warning that prints that
// Location reads as an off-site redirect.
func (spider *Spider) HeadNoRedirects(requestURL url.URL) (*http.Response, error) {
	sameHost := false
	client := &http.Client{
		Transport: &decompressingTransport{
			Transport: http.DefaultTransport,
		},
		Jar:     spider.cookieJar,
		Timeout: spider.requestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			location := ""
			if req != nil && req.Response != nil {
				location = req.Response.Header.Get("Location")
			}
			var from *url.URL
			if len(via) > 0 && via[len(via)-1] != nil {
				from = via[len(via)-1].URL
			}
			var to *url.URL
			if req != nil {
				to = req.URL
			}
			sameHost = sameHostRedirect(from, to)
			if !sameHost {
				log.Warn().Msgf("Redirected %v", location)
			}
			return &ErrRedirect{Location: location}
		},
	}
	req, err := http.NewRequest("HEAD", requestURL.String(), nil)
	if err != nil {
		log.Info().Msgf("Failed to Request %s: %e", requestURL.String(), err)
		return nil, err
	}

	req.Header.Set("User-Agent", spider.userAgent)

	resp, err := client.Do(req)
	if err != nil {
		// The request URL is the other half of the redirect log line. Skip it
		// for a same-host canonical hop; that hop is not a failed request.
		if !sameHost {
			log.Info().Msgf("Failed to Request %s", requestURL.String())
		}
		return nil, err
	}
	return resp, nil
}

// Make a HEAD request to the URL
func (spider *Spider) Head(url url.URL) (*http.Response, int64, error) {
	if !spider.CanCrawl(url) {
		log.Error().Msg("Site blocked by robots.txt")
		return nil, 0, errors.New("URL blocked by robots.txt")
	}
	spider.WaitCrawlDelay()

	client := &http.Client{
		Transport: &decompressingTransport{
			Transport: http.DefaultTransport,
		},
		Jar:     spider.cookieJar,
		Timeout: spider.requestTimeout,
	}
	req, err := http.NewRequest("HEAD", url.String(), nil)
	if err != nil {
		log.Warn().Msgf("Failed to Request %s: %e", url.String(), err)
		return nil, 0, err
	}

	req.Header.Set("User-Agent", spider.userAgent)

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Warn().Msgf("Failed to Request %s", url.String())
		return nil, 0, err
	}
	ms := time.Now().Sub(start).Milliseconds()
	return resp, ms, nil
}

// EndpointExists makes a HEAD request to make sure the URL exists and returns 200
func (spider *Spider) EndpointExists(url url.URL) bool {

	if url.Path == "" || url.Path == "/" {
		log.Info().Msgf("URL is root %s", url.String())
		return false
	}

	client := &http.Client{
		Transport: &decompressingTransport{
			Transport: http.DefaultTransport,
		},
		Jar:     spider.cookieJar,
		Timeout: spider.requestTimeout,
	}
	req, err := http.NewRequest("HEAD", url.String(), nil)
	if err != nil {
		log.Info().Msgf("Failed to Request %s", url.String())
		return false
	}

	req.Header.Set("User-Agent", spider.userAgent)

	resp, err := client.Do(req)
	if err != nil {
		log.Info().Msgf("Failed to Request %s", url.String())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Info().Msgf("Page status %v: %s", resp.StatusCode, url.String())
		return false
	}

	return true
}
