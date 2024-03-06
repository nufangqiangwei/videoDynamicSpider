package models

import (
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/utils"
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

func getSpiderParam(userName, keyName string) (string, error) {
	var userId int64
	if userName == "default" {
		userId = defaultUserId
	} else {
		a := Author{}
		GormDB.Model(&a).First(&a, "author_name = ?", userName)
		if a.Id == 0 {
			return "", nil
		}
		userId = a.Id
	}
	bsh := &BiliSpiderHistory{}
	tx := GormDB.First(bsh, "author_id = ? and key_name = ?", userId, keyName)
	if tx.Error != nil {
		return "", tx.Error
	}
	if tx.RowsAffected == 0 {
		createErr := GormDB.Create(&BiliSpiderHistory{KeyName: keyName, Values: "", AuthorId: userId, LastUpdateTime: time.Now()}).Error
		if createErr != nil {
			return "", createErr
		}
	}
	return bsh.Values, nil
}

func saveSpiderParam(userName, keyName, values string) error {
	var userId int64
	if userName == "default" {
		userId = defaultUserId
	} else {
		a := Author{}
		GormDB.Model(&a).First(&a, "author_name = ?", userName)
		if a.Id == 0 {
			return nil
		}
		userId = a.Id
	}
	return GormDB.Model(&BiliSpiderHistory{}).Where("author_id = ? and key_name = ?", userId, keyName).Update("values", values).Error
}

// GetDynamicBaseline 获取上次获取动态的最后baseline
func GetDynamicBaseline(userName string) string {
	configValue, err := getSpiderParam(userName, "dynamic_baseline")
	if err != nil {
		return ""
	}
	return configValue

}
func SaveDynamicBaseline(baseline, userName string) {
	err := saveSpiderParam(userName, "dynamic_baseline", baseline)
	if err != nil {
		utils.ErrorLog.Printf("保存dynamic_baseline失败:%s", err.Error())
	}
}

func GetHistoryBaseLine(userName string) string {
	configValue, err := getSpiderParam(userName, "history_baseline")
	if err != nil {
		return ""
	}
	return configValue
}
func SaveHistoryBaseLine(baseline, userName string) {
	err := saveSpiderParam(userName, "history_baseline", baseline)
	if err != nil {
		utils.ErrorLog.Printf("保存dynamic_baseline失败:%s", err.Error())
	}
}
