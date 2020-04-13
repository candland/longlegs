package longlegs

import (
	"log"
	"net/url"
	"strings"
)

func CanonicalizeUrl(url *url.URL) string {
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

func ResolvedURL(base *url.URL, urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Print(err)
		return nil
	}
	if u.IsAbs() {
		return removeFragment(u)
		//fmt.Printf("Using base: %s\n", base.String())
	}
	return removeFragment(base.ResolveReference(u))
}

func removeFragment(u *url.URL) *url.URL {
	u.Fragment = ""
	u.RawQuery = "" // remove query string too, for now
	return u
}
