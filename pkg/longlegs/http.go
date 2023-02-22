package longlegs

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

type ErrRedirect struct {
	Location string
}

func (err *ErrRedirect) Error() string {
	return fmt.Sprintf("Redirected %s", err.Location)
}

// Call: `defer resp.Body.Close()`
func (spider *Spider) Get(url url.URL) (*http.Response, int64, error) {
	client := &http.Client{}
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

func (spider *Spider) HeadNoRedirects(url url.URL) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Warn().Msgf("Redirected %v", req.Response.Header.Get("Location"))
			return &ErrRedirect{Location: req.Response.Header.Get("Location")}
		},
	}
	req, err := http.NewRequest("HEAD", url.String(), nil)
	if err != nil {
		log.Info().Msgf("Failed to Request %s: %e", url.String(), err)
		return nil, err
	}

	req.Header.Set("User-Agent", spider.userAgent)

	resp, err := client.Do(req)
	if err != nil {
		log.Info().Msgf("Failed to Request %s", url.String())
		return nil, err
	}
	return resp, nil
}

// EndpointExists makes a HEAD request to make sure the URL exists and returns 200
func (spider *Spider) EndpointExists(url url.URL) bool {

	if url.Path == "" || url.Path == "/" {
		log.Info().Msgf("URL is root %s", url.String())
		return false
	}

	client := &http.Client{}
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

	if resp.StatusCode != 200 {
		log.Info().Msgf("Page status %v: %s", resp.StatusCode, url.String())
		return false
	}

	return true
}
