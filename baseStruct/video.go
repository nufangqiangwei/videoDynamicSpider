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

type TaskClose struct {
	WebSite string
	Code    int
	Data    []UserBaseLine
}
type UserBaseLine struct {
	UserId      int64
	EndBaseLine string
}
