package models

import (
	"gorm.io/gorm"
	"time"
)

// BiliSpiderHistory b站抓取记录
type BiliSpiderHistory struct {
	Id             int64  `gorm:"primaryKey"`
	userId         int64  `gorm:"index"`
	KeyName        string `gorm:"size:255;uniqueIndex"`
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
		userId = 1
	}
	if userId > 0 {
		tx = GormDB.Model(&BiliSpiderHistory{}).First(bsh, "user_id = ? and key_name = ?", userId, "dynamic_baseline")
	} else {
		tx = GormDB.Model(&BiliSpiderHistory{}).Joins("inner join user on bili_spider_history.user_id = user.id").Where("user.user_name = ? and key_name = ?", userName, "dynamic_baseline").First(bsh)
	}
	if tx.Error != nil {
		return ""
	}
	if tx.RowsAffected == 0 {
		GormDB.Create(&BiliSpiderHistory{KeyName: "dynamic_baseline", Values: ""})
	}
	return bsh.Values

}
func SaveDynamicBaseline(baseline string, userName string) {
	var (
		userId int64
	)
	if userName == "default" {
		userId = 1
	}
	if userId > 0 {
		GormDB.Model(&BiliSpiderHistory{}).Where("user_id = ? and key_name = ?", userId, "dynamic_baseline").Update("values", baseline)
	} else {
		GormDB.Model(&BiliSpiderHistory{}).Joins("inner join user on bili_spider_history.user_id = user.id").Where("user.user_name = ? and key_name = ?", userName, "dynamic_baseline").Update("values", baseline)
	}

}

func GetHistoryBaseLine() string {
	bsh := &BiliSpiderHistory{}
	tx := GormDB.First(bsh, "key_name = ? and user_id=764886", "history_baseline")
	if tx.Error != nil {
		return ""
	}
	if tx.RowsAffected == 0 {
		GormDB.Create(&BiliSpiderHistory{KeyName: "history_baseline", Values: ""})
	}
	return bsh.Values
}
func SaveHistoryBaseLine(baseline string) {
	GormDB.Model(&BiliSpiderHistory{}).Where("key_name = ? and user_id=764886", "history_baseline").Update("values", baseline)
}
