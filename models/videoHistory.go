package models

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"time"
	"videoDynamicAcquisition/log"
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
	if tx.Error != nil && !IsMysqlUniqueErr(tx.Error) {
		log.ErrorLog.Println("VideoHistory插入数据错误", tx.Error.Error())
	}
}

func IsMysqlUniqueErr(err error) bool {
	var mysqlErr *mysql.MySQLError
	mysqlErr = new(mysql.MySQLError)
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
