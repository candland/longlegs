package longlegs

import "github.com/PuerkitoBio/goquery"

func (page Page) parseLinks() Page {
	doc := page.Document

	links := doc.Find("a").Map(func(i int, s *goquery.Selection) string {
		href, exists := s.Attr("href")
		if exists {
			return href
		}
		return ""
	})

	linksUniq := map[string]bool{}
	externalUniq := map[string]bool{}

	for _, u := range links {
		rurl := ResolveURL(page.Url, u)
		rurl = CanonicalizeUrl(page.Url, rurl)

		if rurl != nil {
			if page.Url.Hostname() == rurl.Hostname() && (rurl.Scheme == "http" || rurl.Scheme == "https") {
				linksUniq[rurl.String()] = true
			} else {
				externalUniq[rurl.String()] = true
			}
		}
	}

	page.Links = getKeys(linksUniq)
	page.ExternalLinks = getKeys(externalUniq)

	return page
}

func getKeys(hash map[string]bool) []string {
	keys := []string{}
	for k := range hash {
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}
