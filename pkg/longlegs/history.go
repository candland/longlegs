package longlegs

type HistoryEntry struct {
	Crawled bool `json:"crawled"`
	Refs    int  `json:"refs"`
}

type History map[string]*HistoryEntry
