package longlegs

import (
	"net/url"

	"github.com/rs/zerolog/log"
)

type HistoryEntry struct {
	Crawled bool `json:"crawled"`
	Refs    int  `json:"refs"`
	Level   int  `json:"level"`
	Blocked bool `json:"blocked"`
	HTML    bool `json:"html"`
}

type History map[url.URL]*HistoryEntry

func (spider *Spider) addToHistory(level int, links []string) {

	for _, link := range links {
		url := spider.MakeUrl(link)
		if url == nil {
			continue
		}
		if _, exists := spider.history[*url]; !exists {
			log.Debug().Msgf("Adding %s to history %d level.", link, level+1)
			spider.history[*url] = &HistoryEntry{Crawled: false, Level: level + 1}
		}

		spider.history[*url].Refs++
	}
}

// Next: look for next page
func (site *Spider) next(level int) (url.URL, bool) {
	var (
		next    = url.URL{}
		hasNext = false
	)
	for k := range site.history {
		hist := site.history[k]
		if hist.Crawled == false && hist.Level == level {
			next = k
			hasNext = true
			break
		}
	}
	return next, hasNext
}

func (site *Spider) getStatus() (int, int, int) {
	var left, done, level = 0, 0, 100000

	// calc counts
	for _, hist := range site.history {
		if hist.Crawled && hist.HTML {
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
