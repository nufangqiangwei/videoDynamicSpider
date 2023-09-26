package models

import (
	"database/sql"
	"time"
	"videoDynamicAcquisition/utils"
)

// WebSite 网站列表，网站信息
type WebSite struct {
	Id               int64
	WebName          string
	WebHost          string
	WebAuthorBaseUrl string
	WebVideoBaseUrl  string
	CreateTime       time.Time
}

func (w *WebSite) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS website (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				web_name VARCHAR(255) unique,
				web_host VARCHAR(255), 
				web_author_base_url VARCHAR(255), 
				web_video_base_url VARCHAR(255), 
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP
                                   )`
}

func (w *WebSite) GetOrCreate(db *sql.DB) {
	err := dbLock.Lock()
	if err != nil {
		panic(utils.DBFileLock{S: "数据库被锁"})
	}
	defer dbLock.Unlock()

	r, err := db.Exec("INSERT INTO website (web_name, web_host, web_author_base_url, web_video_base_url) VALUES (?, ?, ?, ?)",
		w.WebName, w.WebHost, w.WebAuthorBaseUrl, w.WebVideoBaseUrl)
	if err == nil {
		w.Id, _ = r.LastInsertId()
		w.CreateTime = time.Now()
	} else if utils.IsUniqueErr(err) {
		queryResult := db.QueryRow("SELECT id,create_time FROM website WHERE web_name = ? and web_host=?", w.WebName, w.WebHost)
		queryResult.Scan(&w.Id, &w.CreateTime)
	} else {
		utils.ErrorLog.Println("插入数据错误", err.Error())
	}
}
