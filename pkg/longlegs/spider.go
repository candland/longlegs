package longlegs

import (
	"errors"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/temoto/robotstxt"
)

type ISite interface {
	Process(*Spider, Page)
	Blocked(*Spider, url.URL)
}

type Spider struct {
	hostname     string
	url          url.URL
	userAgent    string
	history      History
	robots       Robots
	robotsData   *robotstxt.RobotsData
	crawlerGroup *robotstxt.Group
}

func (spider *Spider) Robots() Robots {
	return spider.robots
}

func (spider *Spider) Hostname() string {
	return spider.hostname
}

func (spider *Spider) Url() url.URL {
	return spider.url
}

func (spider *Spider) History() History {
	return spider.history
}

func (spider *Spider) UserAgent() string {
	return spider.userAgent
}

func NewSpider(userAgent string, urlStr string) (*Spider, error) {
	spider := &Spider{userAgent: userAgent}
	spider.setUrl(urlStr)

	spider.robotsData = spider.getRobots()
	spider.crawlerGroup = spider.robotsData.FindGroup(spider.userAgent)

	if !spider.CanCrawl(spider.url) {
		log.Fatal().Msg("Site blocked by robots.txt")
	}

	_, err := spider.HeadNoRedirects(spider.url)
	if err != nil {
		if err, success := errors.Unwrap(err).(*ErrRedirect); success {
			spider.setUrl(err.Location)
		} else {
			return nil, err
		}
	}
	spider.WaitCrawlDelay()

	history := make(History)
	history[spider.url] = &HistoryEntry{Crawled: false}

	spider.history = history

	log.Info().Msgf("Found Crawler Group %s with delay of %d", spider.crawlerGroup.Agent, spider.crawlerGroup.CrawlDelay)
	spider.robots = NewRobots(spider.robotsData)

	return spider, nil
}

func (spider *Spider) setUrl(urlStr string) {
	url, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Invalid Redirected URL: %s", urlStr)
	}
	spider.url = *CanonicalizeUrl(*url, nil)
	spider.hostname = strings.ToLower(url.Hostname())
}
