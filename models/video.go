package models

import (
	"database/sql"
	"time"
	"videoDynamicAcquisition/utils"
)

// Video 视频信息
type Video struct {
	Id         int64
	WebSiteId  int64
	AuthorId   int64
	Title      string
	Desc       string
	Duration   int
	Uuid       string
	Url        string
	CoverUrl   string
	UploadTime time.Time
	BiliOffset string
	CreateTime time.Time
}

func (v *Video) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS video (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				web_site_id INTEGER,
				author_id INTEGER,
				title VARCHAR(255),
				video_desc VARCHAR(255),
				duration INTEGER,
				uuid VARCHAR(255),
				url VARCHAR(255),
				cover_url VARCHAR(255),
				upload_time datetime,
				biliOffset VARCHAR(255),
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			   constraint web_site_author_uuid
        			unique (web_site_id, author_id,uuid)
				)`
}

func (v *Video) Save(db *sql.DB) bool {
	r, err := db.Exec("INSERT INTO video (web_site_id, author_id, title,video_desc,duration,uuid, url, cover_url,biliOffset,upload_time) VALUES (?, ?, ?, ?, ?,?,?,?,?,?)",
		v.WebSiteId, v.AuthorId, v.Title, v.Desc, v.Duration, v.Uuid, v.Url, v.CoverUrl, v.BiliOffset, v.UploadTime)
	if err == nil {
		v.Id, _ = r.LastInsertId()
		v.CreateTime = time.Now()
	}
	if err != nil {
		utils.ErrorLog.Println("插入视频错误: ")
		utils.ErrorLog.Println(err.Error())
		return false
	}
	return true
}
func (v *Video) SaveTrc(db *sql.Tx) bool {
	r, err := db.Exec("INSERT INTO video (web_site_id, author_id, title,video_desc,duration,uuid, url, cover_url,biliOffset,upload_time) VALUES (?, ?, ?, ?, ?,?,?,?,?,?)",
		v.WebSiteId, v.AuthorId, v.Title, v.Desc, v.Duration, v.Uuid, v.Url, v.CoverUrl, v.BiliOffset, v.UploadTime)
	if err == nil {
		v.Id, _ = r.LastInsertId()
		v.CreateTime = time.Now()
	}
	if err != nil {
		utils.ErrorLog.Println("插入视频错误: ")
		utils.ErrorLog.Println(err.Error())
		utils.ErrorLog.Println("video: %v\n", v)
		return false
	}
	return true
}

// select strftime( '%H',upload_time) hour,count(*) from video group by hour;
