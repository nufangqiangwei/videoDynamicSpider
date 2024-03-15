package baseStruct

import (
	"time"
	"videoDynamicAcquisition/models"
)

type FollowInfo struct {
	WebSiteId  int64
	UserId     int64
	AuthorName string
	AuthorUUID string
	Avatar     string
	AuthorDesc string
	FollowTime *time.Time
}

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList(chan<- models.Video, chan<- TaskClose)
	GetSelfInfo(string) AccountInfo
}

type AccountInfo interface {
	AccountName() string
	GetAuthorModel() models.Author
}
