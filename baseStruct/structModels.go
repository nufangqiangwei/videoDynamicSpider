package baseStruct

import (
	"fmt"
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

type DateTime time.Time

func (t DateTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}
func (t DateTime) Unix() int64 {
	return time.Time(t).Unix()
}

type CacheUserCookies struct {
	UserName string `gorm:"userName"`
	Content  string `gorm:"content"`
}

// CookiesFlush 读取cookies接口
type CookiesFlush interface {
	WebSiteList() []string
	UserList(webName string) []CacheUserCookies
	GetUserCookies(webSiteName, userName string) string
	UpdateUserCookies(webSiteName, authorName, cookiesContent, userId string) error
}
