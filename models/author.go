package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"videoDynamicAcquisition/utils"
)

// Author 作者信息
type Author struct {
	Id           int64     `json:"id"`
	WebSiteId    int64     `json:"webSiteId"`
	AuthorWebUid string    `json:"authorWebUid"`
	AuthorName   string    `json:"authorName"`
	Avatar       string    `json:"avatar"` // 头像
	Desc         string    `json:"desc"`   // 简介
	Follow       bool      // 是否关注
	FollowTime   time.Time // 关注时间
	Crawl        bool      // 是否爬取
	CreateTime   time.Time
}

var cacheAuthor map[string]Author

func (a *Author) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS author (
    id INT PRIMARY KEY AUTO_INCREMENT,
    web_site_id INT,
    author_web_uid VARCHAR(255),
    author_name VARCHAR(255),
    avatar VARCHAR(255),
    author_desc VARCHAR(255),
    follow BOOL DEFAULT FALSE NOT NULL,
    follow_time DATETIME DEFAULT NULL,
    crawl BOOL DEFAULT FALSE NOT NULL,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT web_site_author UNIQUE (web_site_id, author_web_uid)
);`
}

func (a *Author) GetOrCreate(db *sql.DB) error {
	key := fmt.Sprintf("%d-%s", a.WebSiteId, a.AuthorWebUid)
	if author, ok := cacheAuthor[key]; ok {
		*a = author
		return nil
	}

	queryResult := db.QueryRow("SELECT Id,author_name,avatar,author_desc,follow,follow_time,crawl,create_time FROM author WHERE web_site_id=? AND author_web_uid = ?", a.WebSiteId, a.AuthorWebUid)
	err := queryResult.Scan(&a.Id, &a.AuthorName, &a.Avatar, &a.Desc, &a.Follow, &a.FollowTime, &a.Crawl, &a.CreateTime)
	if errors.Is(err, sql.ErrNoRows) {
		r, err := db.Exec("INSERT INTO author (web_site_id, author_web_uid,author_name,avatar,author_desc,follow,follow_time,crawl) VALUES (?, ?,?,?,?,?,?,?)",
			a.WebSiteId, a.AuthorWebUid, a.AuthorName, a.Avatar, a.Desc, a.Follow, timeCheck(a.FollowTime), a.Crawl)
		if err != nil {
			return err
		}
		a.Id, _ = r.LastInsertId()
		a.CreateTime = time.Now()
	} else if err != nil {
		return err
	}
	cacheAuthor[key] = *a
	return nil
}

func (a *Author) UpdateOrCreate(db *sql.DB) {
	r := db.QueryRow("SELECT id from author where web_site_id=? and author_web_uid=?", a.WebSiteId, a.AuthorWebUid)
	var authorId int64
	err := r.Scan(&authorId)

	if errors.Is(err, sql.ErrNoRows) {
		a.GetOrCreate(db)
		return
	} else if err != nil {
		utils.ErrorLog.Println(err.Error())
		return
	}
	db.Exec("UPDATE author SET avatar=?,author_desc=?,follow=true,follow_time=? WHERE Id=?", a.Avatar, a.Desc, a.FollowTime, authorId)
	cacheAuthor[fmt.Sprintf("%d-%s", a.WebSiteId, a.AuthorWebUid)] = *a

}

func GetAuthorList(db *sql.DB, webSiteId int) (result []Author) {
	rows, err := db.Query("select * from author where web_site_id=?", webSiteId)
	defer rows.Close()
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

func (a *Author) Get(authorId int64, db *sql.DB) {
	key := fmt.Sprintf("%d-%s", a.WebSiteId, a.AuthorWebUid)
	if author, ok := cacheAuthor[key]; ok {
		*a = author
		return
	}
	r, err := db.Query("select * from author where id=? limit 1", authorId)
	defer r.Close()
	if err != nil {
		utils.ErrorLog.Println(err.Error())
		return
	}
	if r.Next() {
		err = r.Scan(&a.Id, &a.WebSiteId, &a.AuthorWebUid, &a.AuthorName, &a.Avatar, &a.Desc, &a.Follow, &a.CreateTime)
	}
	if err != nil {
		utils.ErrorLog.Println(err.Error())
	}
	cacheAuthor[key] = *a
}
