package longlegs

import (
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/temoto/robotstxt"
)

type Site struct {
	Hostname     string                `json:"id"`
	Url          url.URL               `json:"url"`
	History      History               `json:"history"`
	Robots       Robots                `json:"robots"`
	robotsData   *robotstxt.RobotsData `json:"-"`
	crawlerGroup *robotstxt.Group      `json:"-"`
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

	site := &Site{
		Hostname: hostname,
		Url:      *url,
		History:  history,
	}

	site.robotsData = getRobots(site)
	site.crawlerGroup = site.robotsData.FindGroup(site.UserAgent())

	log.Info().Msgf("Found Crawler Group %s with delay of %d", site.crawlerGroup.Agent, site.crawlerGroup.CrawlDelay)
	site.Robots = NewRobots(site.robotsData)

	return site, nil
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

func (site *Site) MakeUrl(path string) url.URL {
	return *ResolveURL(&site.Url, path)
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
