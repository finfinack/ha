package data

type M5Data struct {
	LastUpdated int64     `json:"lastUpdatedSec"`
	Rooms       []*M5Room `json:"rooms"`
}

type M5Room struct {
	Name        string  `json:"name"`
	Temperature float32 `json:"temp"`
}
