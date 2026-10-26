package longlegs

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestSameHostRedirect(t *testing.T) {
	from := mustURL(t, "https://seed.example/path?q=1")

	assert.True(t, sameHostRedirect(from, mustURL(t, "https://seed.example/other")))
	assert.True(t, sameHostRedirect(from, mustURL(t, "https://SEED.EXAMPLE/other")))
	assert.True(t, sameHostRedirect(from, mustURL(t, "http://seed.example/other")))
	assert.True(t, sameHostRedirect(from, mustURL(t, "https://seed.example:443/other")))
	assert.False(t, sameHostRedirect(from, mustURL(t, "https://www.seed.example/")))
	assert.False(t, sameHostRedirect(from, mustURL(t, "https://evil.example.net/")))
	assert.False(t, sameHostRedirect(nil, from))
	assert.False(t, sameHostRedirect(from, nil))
}

func TestHeadNoRedirectsSameHostCanonicalDoesNotWarn(t *testing.T) {
	var buf bytes.Buffer
	restore := captureLog(&buf)
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			http.Redirect(w, r, "/canonical?url=%2F%2Fevil.example.net", http.StatusMovedPermanently)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	requestURL := mustURL(t, server.URL+"/?url=//evil.example.net")
	spider := &Spider{userAgent: "test", requestTimeout: 5 * time.Second}
	_, err := spider.HeadNoRedirects(*requestURL)

	var redirected *ErrRedirect
	assert.True(t, errors.As(err, &redirected), "err = %v", err)
	assert.Equal(t, "/canonical?url=%2F%2Fevil.example.net", redirected.Location)
	assert.NotContains(t, buf.String(), `"level":"warn"`)
	assert.NotContains(t, buf.String(), "Redirected")
	assert.NotContains(t, buf.String(), "Failed to Request")
	assert.NotContains(t, buf.String(), "evil.example.net")
}

func TestHeadNoRedirectsCrossHostStillWarns(t *testing.T) {
	var buf bytes.Buffer
	restore := captureLog(&buf)
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://other.example/landed", http.StatusMovedPermanently)
	}))
	defer server.Close()

	requestURL := mustURL(t, server.URL+"/")
	spider := &Spider{userAgent: "test", requestTimeout: 5 * time.Second}
	_, err := spider.HeadNoRedirects(*requestURL)

	var redirected *ErrRedirect
	assert.True(t, errors.As(err, &redirected), "err = %v", err)
	assert.Equal(t, "https://other.example/landed", redirected.Location)
	assert.Contains(t, buf.String(), `"level":"warn"`)
	assert.Contains(t, buf.String(), "Redirected")
}

func captureLog(buf *bytes.Buffer) func() {
	prev := log.Logger
	log.Logger = zerolog.New(buf)
	return func() { log.Logger = prev }
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	assert.NoError(t, err)
	return u
}
