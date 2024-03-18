package models

import (
	"errors"
	"fmt"
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

type Follow struct {
	Id         int64      `json:"id" gorm:"primaryKey"`
	WebSiteId  int64      `json:"webSiteId" gorm:"type:bigint(20)"`
	AuthorId   int64      `json:"authorId" gorm:"primaryKey"`
	UserId     int64      `json:"userId" gorm:"primaryKey"`
	FollowTime *time.Time `json:"followTime" gorm:"type:datetime"` // 关注时间
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
		a.Id = auth.Id
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

type FollowRelation struct {
	Id           int64
	WebSiteId    int64
	AuthorId     int64
	UserId       int64
	FollowTime   *time.Time
	AuthorWebUid string
}

// GetFollowList 查询指定网站的关注信息，返回的信息结构为map[用户id]map[AuthorWebUid]FollowId
func GetFollowList(webSiteId int64) (result map[int64]map[string]int64) {
	queryData := []FollowRelation{}
	tx := GormDB.Table("follow").Select("follow.Id,follow.web_site_id,follow.author_id,follow.user_id,follow.follow_time,author.author_web_uid").Where("follow.web_site_id=?", webSiteId).Joins(
		"inner join author on author.id=follow.author_id",
	).Scan(&queryData)
	if tx.Error != nil {
		return
	}
	result = make(map[int64]map[string]int64)
	for _, v := range queryData {
		_, ok := result[v.UserId]
		if !ok {
			result[v.UserId] = make(map[string]int64)
		}
		result[v.UserId][v.AuthorWebUid] = v.Id

	}
	return
}

func GetCrawlAuthorList(webSiteId int64) (result []Author) {
	tx := GormDB.Where("crawl=? and web_site_id=?", true, webSiteId).Find(&result)
	if tx.Error != nil {
		return
	}
	return
}

func GetAuthorId(authorName string) (int64, error) {
	if authorName == "default" {
		return defaultUserId, nil
	}
	a := Author{}
	tx := GormDB.Model(&a).First(&a, "author_name = ?", authorName)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return a.Id, nil
}

type AuthorNotExist struct {
	authorName string
}

func (a AuthorNotExist) Error() string {
	return fmt.Sprintf("作者%s不存在", a.authorName)
}
func NewAuthorNotExist(authorName string) AuthorNotExist {
	return AuthorNotExist{authorName: authorName}
}
