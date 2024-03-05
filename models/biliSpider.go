package models

import (
	"gorm.io/gorm"
	"time"
)

const (
	defaultUserId int64 = 764886
)

// BiliSpiderHistory b站抓取记录
type BiliSpiderHistory struct {
	Id             int64  `gorm:"primaryKey"`
	AuthorId       int64  `gorm:"index"`
	KeyName        string `gorm:"size:255;"`
	Values         string `gorm:"size:255"`
	LastUpdateTime time.Time
}

func (m *BiliSpiderHistory) BeforeUpdate(tx *gorm.DB) (err error) {
	m.LastUpdateTime = time.Now()
	return nil
}

// GetDynamicBaseline 获取上次获取动态的最后baseline
func GetDynamicBaseline(userName string) string {
	bsh := &BiliSpiderHistory{}
	var (
		userId int64
		tx     *gorm.DB
	)
	if userName == "default" {
		userId = defaultUserId
	} else {
		a := Author{}
		GormDB.Model(&a).First(&a, "author_name = ?", userName)
		if a.Id == 0 {
			return "查不到用户"
		}
		userId = a.Id
	}
	tx = GormDB.Model(&BiliSpiderHistory{}).First(bsh, "author_id = ? and key_name = ?", userId, "dynamic_baseline")
	if tx.Error != nil {
		return tx.Error.Error()
	}
	if tx.RowsAffected == 0 {
		createErr := GormDB.Create(&BiliSpiderHistory{KeyName: "dynamic_baseline", Values: "", AuthorId: userId, LastUpdateTime: time.Now()}).Error
		if createErr != nil {
			return createErr.Error()
		}
	}
	return bsh.Values

}
func SaveDynamicBaseline(baseline string, userName string) {
	var (
		userId int64
	)
	if userName == "default" {
		userId = defaultUserId
	} else {
		a := Author{}
		GormDB.Model(&a).First(&a, "author_name = ?", userName)
		if a.Id == 0 {
			return
		}
		userId = a.Id
	}
	GormDB.Model(&BiliSpiderHistory{}).Where("author_id = ? and key_name = ?", userId, "dynamic_baseline").Update("values", baseline)
}

func GetHistoryBaseLine() string {
	bsh := &BiliSpiderHistory{}
	tx := GormDB.First(bsh, "key_name = ?", "history_baseline")
	if tx.Error != nil {
		return ""
	}
	if tx.RowsAffected == 0 {
		GormDB.Create(&BiliSpiderHistory{KeyName: "history_baseline", Values: ""})
	}
	return bsh.Values
}
func SaveHistoryBaseLine(baseline string) {
	GormDB.Model(&BiliSpiderHistory{}).Where("key_name = ?", "history_baseline").Update("values", baseline)
}
