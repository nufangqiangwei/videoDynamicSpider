package models

import (
	_ "github.com/mattn/go-sqlite3"
	"time"
	"videoDynamicAcquisition/utils"
)

var (
	dbLock *utils.Flock
)

func init() {
	cacheWebSite = make(map[string]WebSite)
	cacheAuthor = make(map[string]Author)
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
