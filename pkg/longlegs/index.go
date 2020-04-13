package longlegs

import (
	"log"
)

type ProcessPage func(page Page) Page

func Pipeline(fns ...ProcessPage) ProcessPage {
	return func(page Page) Page {
		for _, fn := range fns {
			page = fn(page)
		}
		return page
	}
}

// IndexSite crawls a site and builds the index.
func (site Site) Index(indexLimit int, processPage ProcessPage) Site {
	var (
		done = 0
		left = 0
	)

	// calc counts
	for k := range site.History {
		if site.History[k].Crawled {
			done++
		} else {
			left++
		}
	}

	nextUrl, hasNext := site.Next()
	if hasNext {
		log.Printf("Indexing page %s\n", nextUrl)
		page := ParsePage(site.Hostname, nextUrl)
		site.History[nextUrl].Crawled = true

		if page.Error == nil {
			processPage(page)

			for _, link := range page.Links {
				if _, exists := site.History[link]; !exists {
					log.Printf("Adding %s to history.", link)
					site.History[link] = &HistoryEntry{Crawled: false}
				}

				site.History[link].Refs++
			}
		} else {
			panic(page.Error)
		}
	}

	log.Printf("Indexed %d with %d remaining max of %d.\n", done, left, indexLimit)

	if hasNext && done < indexLimit {
		site = site.Index(indexLimit, processPage)
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
