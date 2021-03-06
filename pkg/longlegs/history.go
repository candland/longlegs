package longlegs

type HistoryEntry struct {
	Crawled bool `json:"crawled"`
	Refs    int  `json:"refs"`
	Level   int  `json:"level"`
}

type History map[string]*HistoryEntry
