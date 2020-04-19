package longlegs

import (
	"net/url"

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
		page := NewPageFromUrl(site, nextUrl)
		site.GetHistory()[nextUrl].Crawled = true

		if page.Error == nil {
			site.Process(page)

			for _, link := range page.Links {
				if _, exists := site.GetHistory()[link]; !exists {
					log.Debug().Msgf("Adding %s to history %d level.", link, level+1)
					site.GetHistory()[link] = &HistoryEntry{Crawled: false, Level: level + 1}
				}

				site.GetHistory()[link].Refs++
			}
		} else {
			log.Warn().Err(page.Error).Msgf("Skipping %s", nextUrl)
		}
	}

	left, done, level = site.GetStatus()

	log.Info().Msgf("Indexed %d with %d remaining max of %d depth of %d", done, left, indexLimit, depth)

	if hasNext && done < indexLimit && level <= depth {
		Index(site, depth, indexLimit)
	}
	return site
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
