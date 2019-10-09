package cue

type Event struct {
	Filter  string `json:"filter"`
	Entries []struct {
		PublishDate string `json:"publishDate"`
	} `json:"entries"`
}
