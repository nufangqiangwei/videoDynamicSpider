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
		db.Exec(baseModel.CreateTale())
	}
	db.Close()
}
