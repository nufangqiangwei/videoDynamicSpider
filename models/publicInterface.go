package models

type VideoCollection interface {
	GetWebSiteName() WebSite
	GetVideoList(chan<- Video, chan<- TaskClose)
	GetSelfInfo(string) AccountInfo
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
	UserId      int64
	EndBaseLine string
}
