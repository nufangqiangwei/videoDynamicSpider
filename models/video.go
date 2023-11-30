package models

import (
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/utils"
)

// Video 视频信息
type Video struct {
	Id               int64 `json:"id" gorm:"primary_key"`
	WebSiteId        int64
	AuthorId         int64
	Title            string
	VideoDesc        string
	Duration         int
	Uuid             string
	Url              string
	CoverUrl         string
	UploadTime       *time.Time
	CreateTime       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	IsMultiplePeople bool      `gorm:"default:false"`
}

func (v *Video) Save() bool {
	tx := GormDB.Create(v)
	if tx.Error != nil {
		utils.ErrorLog.Println("保存视频错误: ")
		utils.ErrorLog.Println(tx.Error.Error())
		return false
	}
	return true
}

func (v *Video) GetByUid(uid string) {
	tx := GormDB.Where("uuid = ?", uid).First(v)
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		utils.ErrorLog.Println("获取视频错误: ")
		utils.ErrorLog.Println(tx.Error.Error())
	}

}
