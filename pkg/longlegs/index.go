package longlegs

import (
	"log"
	"net/url"
)

type IIndex interface {
	Next() (string, bool)
	Process(Page) IIndex
	GetHostname() string
	GetUrl() url.URL
	GetHistory() History
}

func (site Site) Process(page Page) IIndex {
	log.Println("Default page processor. Override to process your pages.")
	return site
}

// IndexSite crawls a site and builds the index.
func Index(site IIndex, indexLimit int) IIndex {
	var (
		done = 0
		left = 0
	)

	// calc counts
	for k := range site.GetHistory() {
		if site.GetHistory()[k].Crawled {
			done++
		} else {
			left++
		}
	}

	nextUrl, hasNext := site.Next()
	if hasNext {
		log.Printf("Indexing page %s\n", nextUrl)
		page := NewPageFromUrl(nextUrl)
		site.GetHistory()[nextUrl].Crawled = true

		if page.Error == nil {
			log.Printf("TYPE: %T\n", site)
			site = site.Process(page) // TODO FIX THIS

			for _, link := range page.Links {
				if _, exists := site.GetHistory()[link]; !exists {
					log.Printf("Adding %s to history.", link)
					site.GetHistory()[link] = &HistoryEntry{Crawled: false}
				}

				site.GetHistory()[link].Refs++
			}
		} else {
			panic(page.Error)
		}
	}

	log.Printf("Indexed %d with %d remaining max of %d.\n", done, left, indexLimit)

	if hasNext && done < indexLimit {
		site = Index(site, indexLimit)
	}
	return site
}

// Next: look for next page
func (site Site) Next() (string, bool) {
	var (
		next    = ""
		hasNext = false
	)
	for k := range site.History {
		if site.History[k].Crawled == false {
			next = k
			hasNext = true
			break
		}
	}
	return next, hasNext
}
