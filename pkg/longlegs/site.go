package longlegs

import (
	"log"
	"net/url"
)

type Site struct {
	Hostname string  `json:"id"`
	Url      url.URL `json:"url"`
	History  History `json:"history"`
}

func NewSite(urlStr string) (Site, error) {
	url, err := url.Parse(urlStr) // TODO canonicalize
	if err != nil {
		log.Printf("Invalid URL: %s\n", urlStr)
		log.Fatal(err)
		return Site{}, err
	}
	hostname := url.Hostname()
	log.Printf("Index %s\n", hostname)

	history := make(History)
	history[urlStr] = &HistoryEntry{Crawled: false}

	return Site{
		Hostname: hostname,
		Url:      *url,
		History:  history,
	}, nil
}

func (site Site) GetHostname() string {
	return site.Hostname
}

func (site Site) GetUrl() url.URL {
	return site.Url
}

func (site Site) GetHistory() History {
	return site.History
}
