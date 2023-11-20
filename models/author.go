package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// Author 作者信息
type Author struct {
	Id           int64      `json:"id" gorm:"primaryKey"`
	WebSiteId    int64      `json:"webSiteId" gorm:"type:bigint(20)"`
	AuthorWebUid string     `json:"authorWebUid" gorm:"type:varchar(255);uniqueIndex"`
	AuthorName   string     `json:"authorName" gorm:"type:varchar(255)"`
	Avatar       string     `json:"avatar" gorm:"type:varchar(255)"` // 头像
	AuthorDesc   string     `json:"desc" gorm:"type:varchar(255)"`   // 简介
	Follow       bool       `gorm:"type:tinyint"`                    // 是否关注
	FollowTime   *time.Time `gorm:"type:datetime"`                   // 关注时间
	Crawl        bool       `gorm:"type:tinyint"`                    // 是否爬取
	CreateTime   time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
}

var cacheAuthor map[string]Author

func (a *Author) GetOrCreate() error {
	key := fmt.Sprintf("%d-%s", a.WebSiteId, a.AuthorWebUid)
	if author, ok := cacheAuthor[key]; ok {
		*a = author
		return nil
	}
	result := gormDB.FirstOrCreate(a, &Author{WebSiteId: a.WebSiteId, AuthorWebUid: a.AuthorWebUid})
	if result.Error != nil {
		return result.Error
	}

	cacheAuthor[key] = *a
	return nil
}

func (a *Author) UpdateOrCreate() error {
	auth := &Author{}
	result := gormDB.Where(Author{WebSiteId: a.WebSiteId, AuthorWebUid: a.AuthorWebUid}).First(auth)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}

	if result.Error == gorm.ErrRecordNotFound {
		// 创建新用户
		result := gormDB.Create(a)
		if result.Error != nil {
			return result.Error
		}
	} else {
		// 更新现有用户
		result := gormDB.Model(auth).Updates(a)
		if result.Error != nil {
			return result.Error
		}
	}

	cacheAuthor[fmt.Sprintf("%d-%s", a.WebSiteId, a.AuthorWebUid)] = *a
	return nil

}

func GetAuthorList(webSiteId int) (result []Author) {
	tx := gormDB.Where("web_site_id=?", webSiteId).Find(&result)
	if tx.Error != nil {
		return
	}
	return
}

func (a *Author) Get(authorId int64) {
	key := fmt.Sprintf("%d-%s", a.WebSiteId, a.AuthorWebUid)
	if author, ok := cacheAuthor[key]; ok {
		*a = author
		return
	}
	tx := gormDB.First(a, authorId)
	if tx.Error != nil {
		return
	}

	cacheAuthor[key] = *a
}

func GetFollowList(webSiteId int64) (result []Author) {
	tx := gormDB.Where("follow=? and web_site_id=?", true, webSiteId).Find(&result)
	if tx.Error != nil {
		return
	}
	return
}
