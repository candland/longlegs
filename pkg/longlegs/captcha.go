package longlegs

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

// CaptchaSolver is implemented by types that can solve CDN bot-protection
// challenges. When longlegs detects a meta-refresh redirect in a page body,
// it follows the redirect, and if the resulting page is a captcha challenge,
// the solver is invoked to solve it. The solver submits the solution via HTTP,
// and the spider's cookie jar captures any cookies set by the server.
type CaptchaSolver interface {
	// CanSolve returns true if the solver recognizes the given page body as
	// a captcha challenge it can handle.
	CanSolve(body []byte, headers http.Header) bool
	// Solve attempts to solve the captcha challenge. It should submit the
	// solution via the spider's HTTP client so the cookie jar captures any
	// resulting cookies. Returns nil on success.
	Solve(spider *Spider, body []byte, headers http.Header) error
}

var metaRefreshRegex = regexp.MustCompile(`(?i)^\s*(\d+)\s*;\s*(?:url\s*=\s*)?['"]?(.+?)['"]?\s*$`)

// parseMetaRefresh extracts the refresh URL from a <meta http-equiv="refresh">
// tag in the given HTML body. Only immediate redirects (delay <= 5 seconds)
// are followed to avoid following delayed refreshes on real pages. Returns
// the refresh URL and true if found, or empty string and false otherwise.
func parseMetaRefresh(body []byte) (string, bool) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", false
	}

	var refreshURL string
	var found bool
	doc.Find(`meta[http-equiv="refresh"], meta[http-equiv="Refresh"]`).Each(func(i int, s *goquery.Selection) {
		if found {
			return
		}
		content, exists := s.Attr("content")
		if !exists {
			return
		}
		match := metaRefreshRegex.FindStringSubmatch(content)
		if match == nil {
			return
		}
		delay, err := strconv.Atoi(match[1])
		if err != nil || delay > 5 {
			return
		}
		refreshURL = match[2]
		found = true
	})

	return refreshURL, found
}

// resolveRefreshURL resolves a potentially relative or encoded refresh URL
// against the page's base URL. Handles URL-encoded paths like %2F.
func resolveRefreshURL(refreshURL string, baseURL *url.URL) (*url.URL, error) {
	// Decode URL-encoded paths (e.g. %2F → /)
	decoded := refreshURL
	if strings.Contains(decoded, "%") {
		if d, err := url.PathUnescape(decoded); err == nil {
			decoded = d
		}
	}

	// Parse the decoded URL
	parsed, err := url.Parse(decoded)
	if err != nil {
		return nil, err
	}

	// Resolve against base URL
	return baseURL.ResolveReference(parsed), nil
}

// handleMetaRefresh checks if the page body contains a meta-refresh redirect.
// If it does, it follows the redirect, optionally solves captcha challenges,
// and returns the final body, status code, and headers. If no meta-refresh is
// found, it returns the original values unchanged.
func (spider *Spider) handleMetaRefresh(pageURL url.URL, body []byte, statusCode int, headers http.Header) ([]byte, int, http.Header) {
	refreshURL, found := parseMetaRefresh(body)
	if !found {
		return body, statusCode, headers
	}

	resolved, err := resolveRefreshURL(refreshURL, &pageURL)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to resolve meta-refresh URL")
		return body, statusCode, headers
	}

	log.Info().Str("from", pageURL.String()).Str("to", resolved.String()).Msg("Following meta-refresh redirect")

	// Follow the meta-refresh
	resp, _, err := spider.Get(*resolved)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to follow meta-refresh redirect")
		return body, statusCode, headers
	}
	defer resp.Body.Close()

	redirectBody, err := readBody(resp)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read meta-refresh response body")
		return body, statusCode, headers
	}

	// Check if this is a captcha challenge
	if spider.captchaSolver != nil && spider.captchaSolver.CanSolve(redirectBody, resp.Header) {
		log.Info().Str("url", resolved.String()).Msg("Captcha challenge detected, solving...")
		if err := spider.captchaSolver.Solve(spider, redirectBody, resp.Header); err != nil {
			log.Warn().Err(err).Msg("Captcha solving failed")
			return body, statusCode, headers
		}
		log.Info().Msg("Captcha solved, re-requesting original URL")

		// Re-request the original URL — cookie jar now has the captcha cookie
		resp2, _, err := spider.Get(pageURL)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to re-request after captcha")
			return body, statusCode, headers
		}
		defer resp2.Body.Close()

		finalBody, err := readBody(resp2)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to read post-captcha body")
			return body, statusCode, headers
		}

		// Check for another meta-refresh (captcha solution may redirect)
		if refreshURL2, found2 := parseMetaRefresh(finalBody); found2 {
			resolved2, err := resolveRefreshURL(refreshURL2, &pageURL)
			if err == nil {
				resp3, _, err := spider.Get(*resolved2)
				if err == nil {
					defer resp3.Body.Close()
					if finalBody2, err := readBody(resp3); err == nil {
						return finalBody2, resp3.StatusCode, resp3.Header
					}
				}
			}
		}

		return finalBody, resp2.StatusCode, resp2.Header
	}

	// Non-captcha meta-refresh — use the redirected content
	return redirectBody, resp.StatusCode, resp.Header
}

// readBody reads and returns the full response body as a byte slice.
func readBody(resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}
