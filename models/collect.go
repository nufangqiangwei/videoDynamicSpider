package models

import (
	"database/sql"
	"time"
	"videoDynamicAcquisition/utils"
)

type Collect struct {
	Id   int64  `json:"id"`
	Type int    `json:"type"`  // 1: 收藏夹 2: 专栏
	BvId int64  `json:"bv_id"` // 收藏夹的bv号
	Name string `json:"name"`  // 收藏夹的名字
}
type CollectVideo struct {
	CollectId int64     `json:"collect_id"`
	VideoId   int64     `json:"video_id"`
	Mtime     time.Time `json:"mtime"`
}

func (ci *Collect) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS collect (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type INTEGER NOT NULL,
		bv_id INTEGER NOT NULL unique ,
		name VARCHAR(255) NOT NULL
	);`
}

func (ci *Collect) CreateOrQuery(db *sql.DB) bool {
	r, err := db.Exec("INSERT INTO collect (type, bv_id, `name`) VALUES ( ?, ?, ?)", ci.Type, ci.BvId, ci.Name)
	if err == nil {
		ci.Id, _ = r.LastInsertId()
		return true
	} else if !utils.IsUniqueErr(err) {
		println(err.Error())
		r, err := db.Query("select id from collect where bv_id = ?", ci.BvId)
		defer r.Close()
		if err != nil {
			return false
		}
		if r.Next() {
			r.Scan(&ci.Id)
		}
		return false
	}
	return true
}

func (ci CollectVideo) CreateTale() string {
	return `CREATE TABLE IF NOT EXISTS collect_video (
		collect_id INTEGER NOT NULL,
		video_id INTEGER NOT NULL,
		mtime datetime,
		    			   constraint web_site_author
				unique (collect_id, video_id)
	);`
}

func (ci CollectVideo) Save(db *sql.DB) {
	_, err := db.Exec("INSERT INTO collect_video (collect_id, video_id,mtime) VALUES (?, ?,?)", ci.CollectId, ci.VideoId, timeCheck(ci.Mtime))
	if err != nil && !utils.IsMysqlUniqueErr(err) {
		utils.ErrorLog.Println("插入数据错误", err.Error())
	}
}

//truncate table author;
//truncate table bili_spider_history;
//truncate table collect;
//truncate table collect_video;
//truncate table video;
//truncate table video_history;
//truncate table website;
