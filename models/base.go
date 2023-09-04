package models

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type BaseModel interface {
	CreateTale() string
}

func InitDB(sqliteDaPath string) {
	println(sqliteDaPath)
	db, err := sql.Open("sqlite3", sqliteDaPath)
	if err != nil {
		panic(err)
	}
	models := []BaseModel{&WebSite{}, &Author{}, &Video{}}
	for _, baseModel := range models {
		_, err = db.Exec(baseModel.CreateTale())
		if err != nil {
			print("创建表失败")
			print(err.Error())
		}
	}
	db.Close()
}
