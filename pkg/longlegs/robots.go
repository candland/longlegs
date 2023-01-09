package longlegs

import (
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

func (site *Site) TestAgent(url string, agent string) bool {
	return site.robotsData.TestAgent(url, agent)
}

func (site *Site) CanCrawl(url string) bool {
	return site.crawlerGroup.Test(url)
}

func (site *Site) CrawlDelay() time.Duration {
	return site.crawlerGroup.CrawlDelay
}

func getRobots(site *Site) *robotstxt.RobotsData {
	robotsUrl := site.MakeUrl("robots.txt")

	page := site.NewRawPageFromUrl(robotsUrl.String())

	log.Info().Msgf("robots.txt\n%s", page.Body)

	robots, err := robotstxt.FromBytes(page.Body)
	if err != nil {
		log.Warn().Err(err).Msgf("No robots.txt, using allow all")
		allowAllRobots, _ := robotstxt.FromBytes([]byte(""))
		return allowAllRobots
	}
	return robots
}
