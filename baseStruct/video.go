package baseStruct

import "time"

type VideoInfo struct {
	WebSite    string
	Title      string
	Desc       string
	Duration   int
	VideoUuid  string
	Url        string
	CoverUrl   string
	AuthorUuid string
	AuthorName string
	AuthorUrl  string
	Baseline   string
	PushTime   time.Time
}
