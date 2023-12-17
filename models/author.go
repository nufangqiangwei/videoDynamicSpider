package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// Author 作者信息
type Author struct {
	Id           int64         `json:"id" gorm:"primaryKey"`
	WebSiteId    int64         `json:"webSiteId" gorm:"type:bigint(20)"`
	AuthorWebUid string        `json:"authorWebUid" gorm:"size:255;uniqueIndex"`
	AuthorName   string        `json:"authorName" gorm:"size:255"`
	Avatar       string        `json:"avatar" gorm:"size:255"`     // 头像
	AuthorDesc   string        `json:"desc" gorm:"size:255"`       // 简介
	Follow       bool          `gorm:"type:tinyint;default:false"` // 是否关注
	FollowTime   *time.Time    `gorm:"type:datetime"`              // 关注时间
	Crawl        bool          `gorm:"type:tinyint;default:false"` // 是否爬取
	CreateTime   time.Time     `gorm:"default:CURRENT_TIMESTAMP(3)"`
	FollowNumber uint64        `gorm:"default:0"` // 关注数
	Videos       []VideoAuthor `gorm:"foreignKey:AuthorId;references:Id"`
}

func (a *Author) GetOrCreate() error {
	result := GormDB.FirstOrCreate(a, &Author{WebSiteId: a.WebSiteId, AuthorWebUid: a.AuthorWebUid})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (a *Author) UpdateOrCreate() error {
	auth := &Author{}
	result := GormDB.Where(Author{WebSiteId: a.WebSiteId, AuthorWebUid: a.AuthorWebUid}).Find(auth)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	if auth.Id <= 0 {
		// 创建新用户
		result := GormDB.Create(a)
		if result.Error != nil {
			return result.Error
		}
	} else {
		// 更新现有用户
		result := GormDB.Model(auth).Updates(a)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil

}

func GetAuthorList(webSiteId int) (result []Author) {
	tx := GormDB.Where("web_site_id=?", webSiteId).Find(&result)
	if tx.Error != nil {
		return
	}
	return
}

func (a *Author) Get(authorId int64) {
	tx := GormDB.First(a, authorId)
	if tx.Error != nil {
		return
	}
}

func GetFollowList(webSiteId int64) (result []Author) {
	tx := GormDB.Where("follow=? and web_site_id=?", true, webSiteId).Find(&result)
	if tx.Error != nil {
		return
	}
	return
}
