package models

import (
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/baseStruct"
)

type UserCookies struct {
	ID         int64     `json:"id" gorm:"primary_key"`
	WebSiteId  int64     `json:"webSiteId" gorm:"index:web_site_id"`
	UserId     int64     `json:"userId" gorm:"index:user_id"` // UserId和AuthorId值为0的时候，代表这个是游客cookies
	AuthorId   int64     `json:"authorId" gorm:"index:author_id"`
	Content    string    `json:"-" gorm:"text"`
	UpdateTime time.Time `json:"updateTime" gorm:"default:CURRENT_TIMESTAMP"`
	Spider     int       `json:"spider" gorm:"default:0"` // 指定哪个爬虫可以读取，0代表所有的都可以读取
	Valid      bool      `json:"valid"`                   // 1有效 0无效
}

// WebSiteCookies 通过数据库实现 baseStruct.CookiesFlush 接口
type WebSiteCookies struct {
	Spider int
}

func (wsc WebSiteCookies) WebSiteList() []string {
	result := make([]string, 0)
	GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id").Where("spider=?", wsc.Spider).Group("web_site.web_name").Pluck("web_site.web_name", &result)
	return result
}
func (wsc WebSiteCookies) UserList(webName string) []baseStruct.CacheUserCookies {
	result := make([]baseStruct.CacheUserCookies, 0)
	GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webName).Joins("inner join author on author.id=author_id").Where("spider=?", wsc.Spider).Select("author.author_name as user_name,content").Scan(&result)
	return result
}
func (wsc WebSiteCookies) GetUserCookies(webSiteName, userName string) string {
	var result UserCookies
	GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webSiteName).Joins("inner join author on author.id=author_id and author.author_name=?", userName).Where("spider=?", wsc.Spider).First(&result)
	return ""
}
func (wsc WebSiteCookies) UpdateUserCookies(webSiteName, authorName, cookiesContent, userId string) error {
	// 判断这条数据是否存在，不存在插入数据，存在才更新
	var (
		result UserCookies
		tx     *gorm.DB
	)
	tx = GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webSiteName).Joins("inner join author on author.id=author_id and author.author_name=?", authorName).First(&result)
	if tx.Error != nil {
		return tx.Error
	}
	if result.ID == 0 {
		// 插入数据，在插入数据前需要判断web_site_id、author_id、userId是否存在
		var webSite WebSite
		tx = GormDB.Model(&WebSite{}).Where("web_name=?", webSiteName).First(&webSite)
		if tx.Error != nil {
			return tx.Error
		}
		if webSite.Id == 0 {
			return NewWebSiteNotExist(webSiteName)
		}
		var author Author
		tx = GormDB.Model(&Author{}).Where("author_name=?", authorName).First(&author)
		if tx.Error != nil {
			return tx.Error
		}
		if author.Id == 0 {
			return NewAuthorNotExist(authorName)
		}
		var user User
		tx = GormDB.Model(&User{}).Where("id=?", userId).First(&user)
		if tx.Error != nil {
			return tx.Error
		}
		if user.ID == 0 {
			return NewUserNotExist(userId)
		}
		tx = GormDB.Create(&UserCookies{WebSiteId: webSite.Id, UserId: user.ID, AuthorId: author.Id, Content: cookiesContent, Spider: wsc.Spider})
		return tx.Error
	}
	tx = GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webSiteName).Joins("inner join author on author.id=author_id and author.author_name=?", authorName).Where("spider=?", wsc.Spider).Update("content", cookiesContent)
	return tx.Error
}
func (wsc WebSiteCookies) UserCookiesInvalid(webSiteName, authorName, cookiesContent, userId string) error {
	return nil
}
func (wsc WebSiteCookies) GetTouristsCookies(webName string) []string {
	return []string{}
}
