package models

import (
	"gorm.io/gorm"
	"time"
)

// BiliSpiderHistory b站抓取记录
type BiliSpiderHistory struct {
	Id             int64  `gorm:"primaryKey"`
	Key            string `gorm:"type:varchar(255);uniqueIndex"`
	Value          string `gorm:"type:varchar(255)"`
	LastUpdateTime time.Time
}

func (m *BiliSpiderHistory) BeforeUpdate(tx *gorm.DB) (err error) {
	m.LastUpdateTime = time.Now()
	return nil
}

// GetDynamicBaseline 获取上次获取动态的最后baseline
func GetDynamicBaseline() string {
	bsh := &BiliSpiderHistory{}

	tx := gormDB.First(bsh, "key = ?", "dynamic_baseline")
	if tx.Error != nil {
		return ""
	}
	if tx.RowsAffected == 0 {
		gormDB.Create(&BiliSpiderHistory{Key: "dynamic_baseline", Value: ""})
	}
	return bsh.Value

}
func SaveDynamicBaseline(baseline string) {
	gormDB.Model(&BiliSpiderHistory{}).Where("key = ?", "dynamic_baseline").Update("value", baseline)

}

func GetHistoryBaseLine() string {
	bsh := &BiliSpiderHistory{}
	tx := gormDB.First(bsh, "key = ?", "history_baseline")
	if tx.Error != nil {
		return ""
	}
	if tx.RowsAffected == 0 {
		gormDB.Create(&BiliSpiderHistory{Key: "history_baseline", Value: ""})
	}
	return bsh.Value
}
func SaveHistoryBaseLine(baseline string) {
	gormDB.Model(&BiliSpiderHistory{}).Where("key = ?", "history_baseline").Update("value", baseline)
}
