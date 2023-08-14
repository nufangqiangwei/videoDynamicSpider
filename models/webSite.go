package models

import (
	"database/sql"
	"time"
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
				id INTEGER PRIMARY KEY AUTO_INCREMENT, 
				web_name VARCHAR(255) UNIQUE,
				web_host VARCHAR(255), 
				web_author_base_url VARCHAR(255), 
				web_video_base_url VARCHAR(255), 
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP
                                   )`
}

func (w *WebSite) GetOrCreate(db *sql.DB) {
	r, err := db.Exec("INSERT INTO website (web_name, web_host, web_author_base_url, web_video_base_url) VALUES (?, ?, ?, ?)",
		w.WebName, w.WebHost, w.WebAuthorBaseUrl, w.WebVideoBaseUrl)
	if err == nil {
		w.Id, _ = r.LastInsertId()
		w.CreateTime = time.Now()
	} else {
		queryResult := db.QueryRow("SELECT * FROM website WHERE web_name = ?", w.WebName)
		queryResult.Scan(&w.Id, &w.CreateTime)
	}
}
