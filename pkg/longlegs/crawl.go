package longlegs

import (
	"time"

	"github.com/rs/zerolog/log"
)

func (spider *Spider) Crawl(site ISite, depth int, limit int) {
	left, done, level := spider.getStatus()

	nextUrl, hasNext := spider.next(level)

	if hasNext {
		log.Info().Msgf("Indexing page %s at %d level", nextUrl.String(), level)

		allowed := spider.CanCrawl(nextUrl)
		if !allowed {
			site.Blocked(spider, nextUrl)
			spider.history[nextUrl].Crawled = true
			spider.history[nextUrl].Blocked = true
		} else {
			page := spider.NewPageFromUrl(nextUrl)
			spider.history[nextUrl].Crawled = true

			if page.Error != nil {
				log.Warn().Err(page.Error).Msgf("Skipping %s", nextUrl.String())
			} else {
				spider.history[nextUrl].HTML = true
				site.Process(spider, page)
				spider.addToHistory(level, page.Links)
			}
		}
	}

	left, done, level = spider.getStatus()
	log.Info().Msgf("Indexed %d at level %d with %d remaining max of %d depth of %d", done, level, left, limit, depth)

	if hasNext && done < limit && level <= depth {
		spider.WaitCrawlDelay()
		spider.Crawl(site, depth, limit)
	}
}

func (spider *Spider) WaitCrawlDelay() {
	delay := spider.CrawlDelay()
	if delay != 0 {
		log.Info().Msgf("Waiting for crawl delay of %d", delay)
		time.Sleep(delay)
	}
}
