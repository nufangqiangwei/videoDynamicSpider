package models

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"videoDynamicAcquisition/utils"
)

var dbLock *utils.Flock

type BaseModel interface {
	CreateTale() string
}

func InitDB(sqliteDaPath string) {
	utils.Info.Println(sqliteDaPath)
	db, err := sql.Open("sqlite3", sqliteDaPath)
	if err != nil {
		panic(err)
	}
	models := []BaseModel{&WebSite{}, &Author{}, &Video{}, &BiliAuthorVideoNumber{}, &BiliSpiderHistory{}, &VideoHistory{}}
	for _, baseModel := range models {
		_, err = db.Exec(baseModel.CreateTale())
		if err != nil {
			utils.ErrorLog.Println("创建表失败")
			utils.ErrorLog.Println(err.Error())
		}
	}
	dbLock = utils.NewFlock(sqliteDaPath)
	db.Close()
}
