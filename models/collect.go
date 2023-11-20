package models

import (
	"strings"
	"time"
	"videoDynamicAcquisition/utils"
)

type Collect struct {
	Id   int64  `json:"id" gorm:"primaryKey"`
	Type int    `json:"type"`  // 1: 收藏夹 2: 专栏
	BvId int64  `json:"bv_id"` // 收藏夹的bv号
	Name string `json:"name"`  // 收藏夹的名字
}
type CollectVideo struct {
	CollectId int64     `json:"collect_id" gorm:"uniqueIndex:collectId_videoId"`
	VideoId   int64     `json:"video_id" gorm:"uniqueIndex:collectId_videoId"`
	Mtime     time.Time `json:"mtime"`
}

func (ci *Collect) Save() bool {
	tx := gormDB.First(ci, "type = ? and bv_id = ?", ci.Type, ci.BvId)
	if tx.Error != nil {
		return false
	}
	if tx.RowsAffected == 0 {
		gormDB.Create(ci)
		return false
	}
	return true
}

func (ci CollectVideo) Save() {
	tx := gormDB.Create(&ci)
	if tx.Error != nil {
		// 如果是唯一索引处突就忽略
		if strings.Contains(tx.Error.Error(), "UNIQUE constraint failed") || strings.Contains(tx.Error.Error(), "Duplicate entry") {
			return
		}
		utils.ErrorLog.Printf("CollectVideo Save error %v\n", tx.Error)
	}
}

type CollectVideoInfo struct {
	CollectId int64
	Video
}

func GetAllCollectVideo() []CollectVideoInfo {
	// select cv.collect_id,v.* from collect_video cv inner join collect c on c.bv_id = cv.collect_id inner join video v on v.id = cv.video_id where c.`type` = 1 and mtime>'0001-01-01 00:00:00+00:00' order by cv.collect_id,mtime desc
	var result []CollectVideoInfo
	gormDB.Table("collect_video cv").
		Select("cv.collect_id,v.*").
		Joins("inner join collect c on c.bv_id = cv.collect_id").
		Joins("inner join video v on v.id = cv.video_id").
		Where("c.`type` = 1 and mtime>'0001-01-01 00:00:00+00:00'").
		Order("cv.collect_id,mtime desc").
		Find(&result)
	return result
}
