package models

import (
	"database/sql"
	"fmt"
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
	UploadTime int64
	CreateTime time.Time
}

func (v *Video) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS video (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				web_site_id INTEGER,
				author_id INTEGER,
				title VARCHAR(255),
				uuid VARCHAR(255) UNIQUE,
				url VARCHAR(255),
				cover_url VARCHAR(255),
				upload_time timestamp,
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			   constraint web_site_author_uuid
        			unique (web_site_id, author_id,uuid)
				)`
}

func (v *Video) Save(db *sql.DB) {

	r, err := db.Exec("INSERT INTO video (web_site_id, author_id, title,uuid, url, cover_url,upload_time) VALUES (?, ?, ?, ?, ?,?,?)",
		v.WebSiteId, v.AuthorId, v.Title, v.Uuid, v.Url, v.CoverUrl, v.CreateTime, v.UploadTime)
	if err == nil {
		v.Id, _ = r.LastInsertId()
		v.CreateTime = time.Now()
	}
	if err != nil {
		println("插入视频错误")
		fmt.Printf("video: %v\n", v)
		println(err.Error())
	}
}
