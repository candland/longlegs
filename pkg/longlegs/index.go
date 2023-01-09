package longlegs

import (
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

type IIndex interface {
	Next(int) (string, bool)
	Process(Page) IIndex
	GetHostname() string
	GetUrl() url.URL
	GetHistory() History
	GetStatus() (int, int, int)
	UserAgent() string
	NewPageFromUrl(string) Page
	NewRawPageFromUrl(string) Page
	MakeUrl(string) url.URL
	TestAgent(string, string) bool
	CanCrawl(string) bool
	CrawlDelay() time.Duration
}

func (site *Site) UserAgent() string {
	return "longlegs/0"
}

func (site *Site) Process(page Page) IIndex {
	log.Warn().Msg("Default page processor. Override to process your pages.")
	return site
}

// IndexSite crawls a site and builds the index.
func Index(site IIndex, depth int, indexLimit int) IIndex {
	left, done, level := site.GetStatus()

	nextUrl, hasNext := site.Next(level)

	if hasNext {
		log.Info().Msgf("Indexing page %s at %d level", nextUrl, level)

		allowed := site.CanCrawl(nextUrl)
		if !allowed {
			log.Warn().Msgf("Blocked by robots, %s", nextUrl)
			site.GetHistory()[nextUrl].Crawled = true
			site.GetHistory()[nextUrl].Blocked = true
		} else {
			page := site.NewPageFromUrl(nextUrl)
			site.GetHistory()[nextUrl].Crawled = true

			if page.Error != nil {
				log.Warn().Err(page.Error).Msgf("Skipping %s", nextUrl)
			} else {
				site.Process(page)
				addNewLinksToHistory(site, page, level)
			}
		}
	}

	left, done, level = site.GetStatus()
	log.Info().Msgf("Indexed %d with %d remaining max of %d depth of %d", done, left, indexLimit, depth)

	if hasNext && done < indexLimit && level <= depth {
		runDelay(site)
		Index(site, depth, indexLimit)
	}

	return site
}

func runDelay(site IIndex) {
	delay := site.CrawlDelay()
	if delay != 0 {
		log.Info().Msgf("Waiting for crawl delay of %d", delay)
		time.Sleep(delay)
	}
}

func addNewLinksToHistory(site IIndex, page Page, level int) {
	for _, link := range page.Links {
		if _, exists := site.GetHistory()[link]; !exists {
			log.Debug().Msgf("Adding %s to history %d level.", link, level+1)
			site.GetHistory()[link] = &HistoryEntry{Crawled: false, Level: level + 1}
		}

		site.GetHistory()[link].Refs++
	}
}

// Next: look for next page
func (site *Site) Next(level int) (string, bool) {
	var (
		next    = ""
		hasNext = false
	)
	for k := range site.History {
		hist := site.History[k]
		if hist.Crawled == false && hist.Level == level {
			next = k
			hasNext = true
			break
		}
	}
	return next, hasNext
}
