package models

import (
	"gorm.io/gorm"
	"time"
)

// BiliSpiderHistory b站抓取记录
type BiliSpiderHistory struct {
	Id             int64  `gorm:"primaryKey"`
	KeyName        string `gorm:"size:255;uniqueIndex"`
	Values         string `gorm:"size:255"`
	LastUpdateTime time.Time
}

func (m *BiliSpiderHistory) BeforeUpdate(tx *gorm.DB) (err error) {
	m.LastUpdateTime = time.Now()
	return nil
}

// GetDynamicBaseline 获取上次获取动态的最后baseline
func GetDynamicBaseline() string {
	bsh := &BiliSpiderHistory{}

	tx := GormDB.First(bsh, "key_name = ?", "dynamic_baseline")
	if tx.Error != nil {
		return ""
	}
	if tx.RowsAffected == 0 {
		GormDB.Create(&BiliSpiderHistory{KeyName: "dynamic_baseline", Values: ""})
	}
	return bsh.Values

}
func SaveDynamicBaseline(baseline string) {
	GormDB.Model(&BiliSpiderHistory{}).Where("key_name = ?", "dynamic_baseline").Update("values", baseline)

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
