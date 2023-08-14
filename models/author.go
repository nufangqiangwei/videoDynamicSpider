package models

import (
	"database/sql"
	"time"
)

// Author 作者信息
type Author struct {
	Id         int64
	WebSiteId  int64
	AuthorName string
	CreateTime time.Time
}

func (a *Author) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS author (
				id INTEGER PRIMARY KEY AUTO_INCREMENT,
				web_site_id INTEGER,
				author_name VARCHAR(255) UNIQUE,
				create_time DATETIME DEFAULT CURRENT_TIMESTAMP
				)`
}

func (a *Author) GetOrCreate(db *sql.DB) {
	r, err := db.Exec("INSERT INTO author (web_site_id, author_name) VALUES (?, ?)", a.WebSiteId, a.AuthorName)
	if err == nil {
		a.Id, _ = r.LastInsertId()
		a.CreateTime = time.Now()
	} else {
		queryResult := db.QueryRow("SELECT * FROM author WHERE author_name = ?", a.AuthorName)
		queryResult.Scan(&a.Id, &a.CreateTime)
	}
}
