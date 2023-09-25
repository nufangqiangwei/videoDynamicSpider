package models

import (
	_ "github.com/mattn/go-sqlite3"
	"videoDynamicAcquisition/utils"
)

var dbLock *utils.Flock

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
