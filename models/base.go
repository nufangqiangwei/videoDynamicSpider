package models

import (
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/utils"
)

var (
	dbLock *utils.Flock
	db     *gorm.DB
)

func InitDB(dsn string) {
	cacheWebSite = make(map[string]WebSite)
	cacheAuthor = make(map[string]Author)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}
}

func timeCheck(t time.Time) interface{} {
	var followTime interface{} = nil
	if t.Unix() > 0 {
		followTime = t.Format("2006-01-02 15:04:05")
	}
	return followTime
}

type BaseModel interface {
	CreateTale() string
}

func CreateDbLock(dbPath string) {
	if dbLock == nil {
		dbLock = utils.NewFlock(dbPath)
	} else {
		println("dbLock已经存在")
	}
}
