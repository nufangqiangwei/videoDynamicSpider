package models

type VideoCollection interface {
	GetWebSiteName() WebSite
	GetVideoList(chan<- Video, chan<- TaskClose)
	GetSelfInfo(string) AccountInfo
	GetVideoInfo(string, bool) Video
	GetHotVideoList(chan Video, chan<- int64)
}

type AccountInfo interface {
	AccountName() string
	GetAuthorModel() Author
}

type TaskClose struct {
	WebSite string
	Code    int
	Data    []UserBaseLine
}

type UserBaseLine struct {
	AuthorId    int64
	EndBaseLine string
}
