package models

import (
	"strings"
	"time"
	"videoDynamicAcquisition/log"
	webSiteGRPC "videoDynamicAcquisition/proto"
)

type Collect struct {
	Id        int64          `json:"id" gorm:"primaryKey"`
	AuthorId  int64          `json:"author_id"`
	Type      int            `json:"type"`                 // 1: 收藏夹 2: 专栏
	BvId      int64          `json:"bv_id"`                // 收藏夹的bv号
	Name      string         `json:"name" gorm:"size:255"` // 收藏夹的名字
	VideoInfo []CollectVideo `gorm:"foreignKey:CollectId;references:Id"`
}
type CollectVideo struct {
	Id        int64      `gorm:"primary_key"`
	CollectId int64      `json:"collect_id" gorm:"uniqueIndex:collectId_videoId"`
	VideoId   int64      `json:"video_id" gorm:"uniqueIndex:collectId_videoId"`
	Mtime     *time.Time `json:"mtime"`
	IsDel     bool       `json:"is_del" gorm:"index:is_del"`
	IsInvalid bool       `json:"is_invalid" gorm:"index:is_invalid"`
}

func (ci *Collect) Save() bool {
	tx := GormDB.First(ci, "type = ? and bv_id = ?", ci.Type, ci.BvId)
	if tx.Error != nil {
		return false
	}
	if tx.RowsAffected == 0 {
		GormDB.Create(ci)
		return false
	}
	return true
}

func (ci CollectVideo) Save() {
	tx := GormDB.Create(&ci)
	if tx.Error != nil {
		// 如果是唯一索引处突就忽略
		if strings.Contains(tx.Error.Error(), "UNIQUE constraint failed") || strings.Contains(tx.Error.Error(), "Duplicate entry") {
			return
		}
		log.ErrorLog.Printf("CollectVideo Save error %v\n", tx.Error)
	}
}

func GetUserCollectList(userId int64) []map[string]interface{} {
	sql := `select sub.bv_id, sub.video_count, bv.video_id, v.title, v.uuid
from (SELECT c.bv_id,
             COUNT(*)   AS video_count,
             MAX(cv.id) AS latest_id
      FROM collect_video cv
               inner join collect c on c.id = cv.collect_id and c.author_id = ?
      GROUP BY collect_id) sub
         left join collect_video bv on bv.id = latest_id
         left join video v on v.id = bv.video_id`
	result := make([]webSiteGRPC.CollectionInfo, 0)
	GormDB.Raw(sql, userId).Scan(&result)
	return result
}

type CollectVideoInfo struct {
	CollectId int64
	Video
}

func GetAllCollectVideo() []CollectVideoInfo {
	// select cv.collect_id,v.* from collect_video cv inner join collect c on c.bv_id = cv.collect_id inner join video v on v.id = cv.video_id where c.`type` = 1 and mtime>'0001-01-01 00:00:00+00:00' order by cv.collect_id,mtime desc
	var result []CollectVideoInfo
	GormDB.Table("collect_video cv").
		Select("cv.collect_id,v.*").
		Joins("inner join collect c on c.bv_id = cv.collect_id").
		Joins("inner join video v on v.id = cv.video_id").
		Where("c.`type` = 1 and mtime>'0001-01-01 00:00:00+00:00'").
		Order("cv.collect_id,mtime desc").
		Find(&result)
	return result
}

func GetAllCollect() []Collect {
	var result []Collect
	GormDB.Find(&result)
	return result
}
