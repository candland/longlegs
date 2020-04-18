package longlegs

import (
	"log"
	"net/url"
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

func (site Site) UserAgent() string {
	return "longlegs v0"
}

func (site Site) Process(page Page) IIndex {
	log.Println("Default page processor. Override to process your pages.")
	return site
}

// IndexSite crawls a site and builds the index.
func Index(site IIndex, depth int, indexLimit int) IIndex {
	left, done, level := site.GetStatus()

	nextUrl, hasNext := site.Next(level)
	if hasNext {
		log.Printf("Indexing page %s at %d level\n", nextUrl, level)
		page := NewPageFromUrl(site, nextUrl)
		site.GetHistory()[nextUrl].Crawled = true

		if page.Error == nil {
			site = site.Process(page)

			for _, link := range page.Links {
				if _, exists := site.GetHistory()[link]; !exists {
					log.Printf("Adding %s to history %d level.", link, level+1)
					site.GetHistory()[link] = &HistoryEntry{Crawled: false, Level: level + 1}
				}

				site.GetHistory()[link].Refs++
			}
		} else {
			panic(page.Error)
		}
	}

	left, done, level = site.GetStatus()

	log.Printf("Indexed %d with %d remaining max of %d depth of %d.\n", done, left, indexLimit, depth)

	if hasNext && done < indexLimit && level <= depth {
		site = Index(site, depth, indexLimit)
	}
	return site
}

// Next: look for next page
func (site Site) Next(level int) (string, bool) {
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
