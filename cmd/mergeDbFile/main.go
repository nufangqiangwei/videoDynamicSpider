package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

// 将 E:\GoCode\videoDynamicAcquisition\videoInfo.db 文件中的video表
// 合并到 E:\GoCode\videoDynamicAcquisition\cmd\spider\videoInfo.db 文件中的video表
const saveDbPath = "E:\\GoCode\\videoDynamicAcquisition\\videoInfo.db"

var sourceDbPath = []string{
	"E:\\GoCode\\videoDynamicAcquisition\\cmd\\spider\\videoInfo.db",
}

func main() {
	// 1. 打开两个数据库文件
	// 2. 从第一个数据库中查询出所有的数据
	// 3. 将第一个数据库中的数据插入到第二个数据库中
	// 4. 关闭两个数据库文件
	saveDb, err := sql.Open("sqlite3", saveDbPath)
	if err != nil {
		fmt.Printf("打开目标数据库失败: %s\n", err.Error())
		return
	}
	for _, s := range sourceDbPath {
		sourceDb, err := sql.Open("sqlite3", s)
		if err != nil {
			fmt.Printf("打开%s源数据库失败: %s\n", s, err.Error())
			continue
		}
		mergeData(sourceDb, saveDb)
		sourceDb.Close()
	}
	saveDb.Close()
}

func mergeData(sourceDataDb *sql.DB, saveDb *sql.DB) {
	var (
		webName         string
		authorWebUid    string
		authorName      string
		authorDesc      string
		avatar          string
		authorId        int64
		videoTitle      string
		cover           string
		Desc            string
		duration        int
		uuid            string
		uploadTime      time.Time
		err             error
		r               sql.Result
		lastWebsiteName string
		lastWebsiteId   int64
		videoId         int64
	)
	db1AuthorRows, err := sourceDataDb.Query("select w.web_name, a.author_web_uid, a.author_name, a.author_desc, a.avatar, v.title, v.cover_url, v.video_desc, v.duration, v.uuid, v.upload_time from video v inner join author a on a.id = v.author_id inner join website w on w.id = v.web_site_id where a.author_web_uid!=0")
	if err != nil {
		fmt.Printf("sourceDataDb查询视频信息数据错误: %v\n", err)
		return
	}
	for db1AuthorRows.Next() {
		err = db1AuthorRows.Scan(&webName, &authorWebUid, &authorName, &authorDesc, &avatar, &videoTitle, &cover, &Desc, &duration, &uuid, &uploadTime)
		if err != nil {
			fmt.Printf("sourceDataDb绑定数据错误: %v\n", err)
			continue
		}
		// 查询website表中是否有该网站
		if lastWebsiteName != webName {
			lastWebsiteName = webName
			err = saveDb.QueryRow("select id from website where web_name = ?", webName).Scan(&lastWebsiteId)
			if err != nil {
				if err == sql.ErrNoRows {
					r, err = saveDb.Exec("INSERT INTO website(web_name) values(?)", webName)
					if err != nil {
						fmt.Printf("%s插入website失败: %s\n", webName, err.Error())
						return
					}
					lastWebsiteId, _ = r.LastInsertId()
				} else {
					fmt.Printf("%s查询website失败: %s\n", webName, err.Error())
					continue
				}
			}
		}
		// 查询author表中是否有该up
		err = saveDb.QueryRow("select id from author where author_web_uid = ?", authorWebUid).Scan(&authorId)
		if err != nil {
			if err == sql.ErrNoRows {
				r, err = saveDb.Exec("INSERT INTO author(web_site_id, author_web_uid, author_name, author_desc, avatar) values(?, ?, ?, ?, ?)", lastWebsiteId, authorWebUid, authorName, authorDesc, avatar)
				if err != nil {
					fmt.Printf("%s插入author失败: %s\n", authorName, err.Error())
					continue
				}
				authorId, _ = r.LastInsertId()
			} else {
				fmt.Printf("%s查询up失败: %s\n", authorName, err.Error())
				continue
			}
		}
		// 查询video表中是否有该视频
		err = saveDb.QueryRow("select id from video where uuid = ?", uuid).Scan(&videoId)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = saveDb.Exec("INSERT INTO video(web_site_id, author_id, title, cover_url, video_desc, duration, uuid, upload_time) values(?, ?, ?, ?, ?, ?, ?, ?)", lastWebsiteId, authorId, videoTitle, cover, Desc, duration, uuid, uploadTime)
				if err != nil {
					fmt.Printf("%s插入video失败: %s\n", videoTitle, err.Error())
					continue
				}
			} else {
				fmt.Printf("%s查询video失败: %s\n", uuid, err.Error())
				continue
			}
		}

	}
	db1AuthorRows.Close()

	db1AuthorRows, err = sourceDataDb.Query("select w.web_name,v.uuid,vh.view_time from video_history vh inner join website w on w.id = vh.web_site_id inner join video v on v.id = vh.video_id where uuid != ''")
	if err != nil {
		fmt.Printf("sourceDataDb查询视频观看历史数据错误: %v\n", err)
		return
	}
	for db1AuthorRows.Next() {
		err = db1AuthorRows.Scan(&webName, &uuid, &uploadTime)
		if err != nil {
			fmt.Printf("sourceDataDb绑定数据错误: %v\n", err)
			continue
		}
		// 查询website表中是否有该网站
		if lastWebsiteName != webName {
			lastWebsiteName = webName
			err = saveDb.QueryRow("select id from website where web_name = ?", webName).Scan(&lastWebsiteId)
			if err != nil {
				fmt.Printf("%s查询website失败: %s\n", webName, err.Error())
				continue
			}
		}
		// 查询video表中是否有该视频
		err = saveDb.QueryRow("select id from video where uuid = ?", uuid).Scan(&videoId)
		if err != nil {
			fmt.Printf("%s查询video失败: %s\n", uuid, err.Error())
			continue
		}
		// 查询video_history表中是否有该视频
		err = saveDb.QueryRow("select id from video_history where video_id = ? and view_time = ?", videoId, uploadTime).Scan(&videoId)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = saveDb.Exec("INSERT INTO video_history(web_site_id, video_id, view_time) values(?, ?, ?)", lastWebsiteId, videoId, uploadTime)
				if err != nil {
					fmt.Printf("%s插入video_history失败: %s\n", uuid, err.Error())
					continue
				}
			} else {
				fmt.Printf("%s查询video_history失败: %s\n", uuid, err.Error())
				continue
			}
		}

	}
}
