package models

import (
	"time"
	"videoDynamicAcquisition/utils"
)

type VideoHistory struct {
	Id        int64 `gorm:"primaryKey"`
	WebSiteId int64
	VideoId   int64
	ViewTime  time.Time
	WebUUID   string `gorm:"size:255"`
	Duration  int    // 视频的观看进度
}

func (vh VideoHistory) Save() {
	tx := GormDB.Create(&vh)
	if tx.Error != nil && !utils.IsMysqlUniqueErr(tx.Error) {
		utils.ErrorLog.Println("VideoHistory插入数据错误", tx.Error.Error())
	}
}
