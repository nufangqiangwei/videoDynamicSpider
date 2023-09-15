package models

import (
	"database/sql"
	"time"
)

type VideoHistory struct {
	Id        int64
	WebSiteId int64
	VideoId   int64
	ViewTime  time.Time
}

func (vh VideoHistory) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS video_history (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				web_site_id INTEGER,
				video_id INTEGER,
				view_time DATETIME DEFAULT CURRENT_TIMESTAMP
				)`
}

func (vh VideoHistory) Save(db *sql.DB) {
	db.Exec("INSERT INTO video_history (web_site_id, video_id, view_time) VALUES (?, ?, ?)", vh.WebSiteId, vh.VideoId, vh.ViewTime)
}
