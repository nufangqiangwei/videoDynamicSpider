package models

import (
	"database/sql"
	"time"
)

// Video 视频信息
type Video struct {
	Id         int64
	WebSiteId  int64
	AuthorId   int64
	Title      string
	Uuid       string
	Url        string
	CoverUrl   string
	CreateTime time.Time
}

func (v *Video) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS video (
				id INTEGER PRIMARY KEY AUTO_INCREMENT,
				web_site_id INTEGER,
				author_id INTEGER,
				title VARCHAR(255),
				uuid VARCHAR(255) UNIQUE,
				url VARCHAR(255),
				cover_url VARCHAR(255),
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP
				)`
}

func (v *Video) Save(db *sql.DB) {

	r, err := db.Exec("INSERT INTO video (web_site_id, author_id, title,uuid, url, cover_url,create_time) VALUES (?, ?, ?, ?, ?,?)",
		v.WebSiteId, v.AuthorId, v.Title, v.Uuid, v.Url, v.CoverUrl, v.CreateTime)
	if err == nil {
		v.Id, _ = r.LastInsertId()
		v.CreateTime = time.Now()
	}
}
