package longlegs

import (
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

type Site struct {
	Hostname string  `json:"id"`
	Url      url.URL `json:"url"`
	History  History `json:"history"`
}

func NewSite(urlStr string) (*Site, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		log.Warn().Err(err).Msgf("Invalid URL: %s", urlStr)
		return &Site{}, err
	}
	url = CanonicalizeUrl(url, nil)
	hostname := strings.ToLower(url.Hostname())
	log.Debug().Msgf("Hostname %s", hostname)

	history := make(History)
	history[urlStr] = &HistoryEntry{Crawled: false}

	return &Site{
		Hostname: hostname,
		Url:      *url,
		History:  history,
	}, nil
}

func (site *Site) GetHostname() string {
	return site.Hostname
}

func (site *Site) GetUrl() url.URL {
	return site.Url
}

func (site *Site) GetHistory() History {
	return site.History
}

func (site *Site) GetStatus() (int, int, int) {
	var left, done, level = 0, 0, 100000

	// calc counts
	for k := range site.GetHistory() {
		hist := site.GetHistory()[k]
		if hist.Crawled {
			done++
		} else {
			left++
			if hist.Level < level {
				level = hist.Level
			}
		}
	}

	return left, done, level
}
