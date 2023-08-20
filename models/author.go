package models

import (
	"database/sql"
	"time"
)

// Author 作者信息
type Author struct {
	Id           int64
	WebSiteId    int64
	AuthorWebUid string
	AuthorName   string
	CreateTime   time.Time
}

func (a *Author) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS author (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				web_site_id INTEGER,
				author_web_uid VARCHAR(255) ,
				author_name VARCHAR(255) ,
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    			   constraint web_site_author
        			unique (web_site_id, author_web_uid)
				)`
}

func (a *Author) GetOrCreate(db *sql.DB) {
	r, err := db.Exec("INSERT INTO author (web_site_id, author_web_uid,author_name) VALUES (?, ?,?)", a.WebSiteId, a.AuthorWebUid, a.AuthorName)
	if err == nil {
		a.Id, _ = r.LastInsertId()
		a.CreateTime = time.Now()
	} else {
		queryResult := db.QueryRow("SELECT Id,create_time FROM author WHERE web_site_id=? AND author_web_uid = ?", a.WebSiteId, a.AuthorWebUid)
		queryResult.Scan(&a.Id, &a.CreateTime)
	}
}
