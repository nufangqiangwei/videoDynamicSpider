package models

import (
	"fmt"
	"time"
)

// WebSite 网站列表，网站信息
type WebSite struct {
	Id               int64     `gorm:"primaryKey"`
	WebName          string    `gorm:"unique;size:255"`
	WebHost          string    `gorm:"size:255"`
	WebAuthorBaseUrl string    `gorm:"size:255"`
	WebVideoBaseUrl  string    `gorm:"size:255"`
	CreateTime       time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
}

var cacheWebSite map[string]WebSite

func (w *WebSite) GetOrCreate() error {
	if website, ok := cacheWebSite[w.WebName]; ok {
		*w = website
		return nil
	}

	result := GormDB.FirstOrCreate(w, &WebSite{WebName: w.WebName})
	if result.Error != nil {
		return result.Error
	}

	cacheWebSite[w.WebName] = *w
	return nil
}

type WebSiteNotExist struct {
	webSiteName string
}

func (w WebSiteNotExist) Error() string {
	return fmt.Sprintf("网站%s不存在", w.webSiteName)
}
