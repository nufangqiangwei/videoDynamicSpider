package models

import (
	"time"
	"videoDynamicAcquisition/utils"
)

type VideoHistory struct {
	Id        int64
	WebSiteId int64
	VideoId   int64
	ViewTime  time.Time
	WebUUID   string
}

func (vh VideoHistory) Save() {
	tx := GormDB.Create(&vh)
	if tx.Error != nil {
		utils.ErrorLog.Println("VideoHistory插入数据错误", tx.Error.Error())
	}
}
