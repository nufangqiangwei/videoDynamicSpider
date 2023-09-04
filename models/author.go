package models

import (
	"database/sql"
	"time"
)

// Author 作者信息
type Author struct {
	Id           int64  `json:"id"`
	WebSiteId    int64  `json:"webSiteId"`
	AuthorWebUid string `json:"authorWebUid"`
	AuthorName   string `json:"authorName"`
	Avatar       string `json:"avatar"` // 头像
	Desc         string `json:"desc"`   // 简介
	Follow       bool   // 是否关注
	CreateTime   time.Time
}

func (a *Author) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS author (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				web_site_id INTEGER,
				author_web_uid VARCHAR(255) ,
				author_name VARCHAR(255) ,
				avatar VARCHAR(255) ,
				author_desc VARCHAR(255) ,
				follow bool default false not null,
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    			   constraint web_site_author
        			unique (web_site_id, author_web_uid)
				)`
}

func (a *Author) GetOrCreate(db *sql.DB) {
	r, err := db.Exec("INSERT INTO author (web_site_id, author_web_uid,author_name,avatar,author_desc) VALUES (?, ?,?,?,?)", a.WebSiteId, a.AuthorWebUid, a.AuthorName, a.Avatar, a.Desc)
	if err == nil {
		a.Id, _ = r.LastInsertId()
		a.CreateTime = time.Now()
	} else {
		avatar := ""
		author_desc := ""
		queryResult := db.QueryRow("SELECT Id,create_time,avatar,author_desc FROM author WHERE web_site_id=? AND author_web_uid = ?", a.WebSiteId, a.AuthorWebUid)
		queryResult.Scan(&a.Id, &a.CreateTime, &avatar, &author_desc)
		if a.Avatar != "" && a.Desc != "" && (a.Avatar != avatar || a.Desc != author_desc) {
			db.Exec("UPDATE author SET avatar=?,author_desc=? WHERE Id=?", a.Avatar, a.Desc, a.Id)
		}
	}
}

func GetAuthorList(db *sql.DB, webSiteId int) (result []Author) {
	rows, err := db.Query("select * from author where web_site_id=?", webSiteId)
	result = make([]Author, 0)
	if err != nil {
		return
	}

	for rows.Next() {
		a := Author{}
		rows.Scan(&a.Id, &a.WebSiteId, &a.AuthorWebUid, &a.AuthorName, &a.Avatar, &a.Desc, &a.Follow, &a.CreateTime)
		result = append(result, a)
	}
	return
}
