package longlegs

import (
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/temoto/robotstxt"
)

type Robots struct {
	Host     string   `json:"host"`
	Sitemaps []string `json:"sitemaps"`
}

func NewRobots(robotsData *robotstxt.RobotsData) Robots {
	return Robots{
		Host:     robotsData.Host,
		Sitemaps: robotsData.Sitemaps,
	}
}

// dep
// func (site *Site) TestAgent(url string, agent string) bool {
//   return site.robotsData.TestAgent(url, agent)
// }
//
// // dep
// func (site *Site) CanCrawl(url string) bool {
//   return site.crawlerGroup.Test(url)
// }
//
// // dep
// func (site *Site) CrawlDelay() time.Duration {
//   return site.crawlerGroup.CrawlDelay
// }

func (spider *Spider) TestAgent(url string) bool {
	return spider.robotsData.TestAgent(url, spider.userAgent)
}

func (spider *Spider) CanCrawl(url url.URL) bool {
	return spider.crawlerGroup.Test(url.String())
}

func (spider *Spider) CrawlDelay() time.Duration {
	return spider.crawlerGroup.CrawlDelay
}

func (spider *Spider) getRobots() *robotstxt.RobotsData {
	url := spider.MakeUrl("robots.txt")
	if url == nil {
		log.Warn().Msgf("Failed to make robots.txt URL")
		allowAllRobots, _ := robotstxt.FromBytes([]byte(""))
		return allowAllRobots
	}

	page := spider.NewRawPageFromUrl(*url)

	log.Info().Msgf("robots.txt\n%s", page.Body)

	robots, err := robotstxt.FromBytes(page.Body)
	if err != nil {
		log.Warn().Err(err).Msgf("No robots.txt, using allow all")
		allowAllRobots, _ := robotstxt.FromBytes([]byte(""))
		return allowAllRobots
	}
	return robots
}
