package longlegs

import (
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

func CanonicalizeUrlStr(url *url.URL) string {
	// //host.lower[/path if not /][?query] - #nofragrment
	can := strings.ToLower(url.Hostname())
	if url.EscapedPath() != "/" {
		can += url.EscapedPath()
	}
	if url.RawQuery != "" {
		can += url.RawQuery
	}
	return can
}

// urlToId removes the Scheme and ://. Should call `CanonicalizeUrl` before calling.
func urlToId(url *url.URL) string {
	can := url.Host
	can += url.EscapedPath()
	if url.RawQuery != "" {
		can += "?" + url.RawQuery
	}
	return can
}

// Try to canonicalize the url. Only changes URLs where the Hostname == the base.Hostname.
func CanonicalizeUrl(base *url.URL, url *url.URL) *url.URL {
	if url == nil {
		url = base
	}

	// Make sure it's an absolute URL
	if !url.IsAbs() {
		url = base.ResolveReference(url)
	}

	// Don't worry about or change external URLs
	if strings.ToLower(base.Hostname()) != strings.ToLower(url.Hostname()) {
		return url
	}

	// Lower host
	url.Host = strings.ToLower(url.Host)

	// Set scheme to base
	url.Scheme = base.Scheme

	// Remove tailing slash if only trailing slash
	if url.EscapedPath() == "/" {
		url.Path = ""
	}

	// Remove any fragments
	url.Fragment = ""

	// Don't include ? without a query
	url.ForceQuery = false

	// Sorted query strings
	url.RawQuery = url.Query().Encode()

	return url
}

// Resolves URL to the base URL if the url isn't already absolute.
func ResolveURL(base *url.URL, urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Info().Err(err).Msgf("Failed to parse %s; returning empty URL", urlStr)
		return &url.URL{}
	}
	if !u.IsAbs() {
		u = base.ResolveReference(u)
	}
	return u
}
