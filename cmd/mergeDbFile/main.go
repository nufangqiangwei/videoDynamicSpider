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
	sourceDBs := make([]*sql.DB, 0)
	for _, s := range sourceDbPath {
		sourceDb, err := sql.Open("sqlite3", s)
		if err != nil {
			fmt.Printf("打开%s源数据库失败: %s\n", s, err.Error())
			continue
		}
		sourceDBs = append(sourceDBs, sourceDb)

	}
	mergeData(sourceDBs, saveDb)
	for _, db_1 := range sourceDBs {
		db_1.Close()
	}
	saveDb.Close()
}

type videoInfo struct {
	webName      string
	authorWebUid string
	authorName   string
	authorDesc   string
	avatar       string
	videoTitle   string
	cover        string
	Desc         string
	duration     int
	uuid         string
	uploadTime   time.Time
}

func getVideoInfo(dbS []*sql.DB, resultChan chan videoInfo) {
	var (
		webName      string
		authorWebUid string
		authorName   string
		authorDesc   string
		avatar       string
		videoTitle   string
		cover        string
		Desc         string
		duration     int
		uuid         string
		uploadTime   time.Time
	)
	for _, db := range dbS {
		db1AuthorRows, err := db.Query("select w.web_name, a.author_web_uid, a.author_name, a.author_desc, a.avatar, v.title, v.cover_url, v.video_desc, v.duration, v.uuid, v.upload_time from video v inner join author a on a.id = v.author_id inner join website w on w.id = v.web_site_id where a.author_web_uid!=0")
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
			resultChan <- videoInfo{
				webName:      webName,
				authorWebUid: authorWebUid,
				authorName:   authorName,
				authorDesc:   authorDesc,
				avatar:       avatar,
				videoTitle:   videoTitle,
				cover:        cover,
				Desc:         Desc,
				duration:     duration,
				uuid:         uuid,
				uploadTime:   uploadTime,
			}
		}
		db1AuthorRows.Close()
	}
	close(resultChan)
}

type videoHistory struct {
	webName  string
	uuid     string
	ViewTime time.Time
}

func getVideoHistory(dbs []*sql.DB, resultChan chan videoHistory) {
	var (
		webName  string
		uuid     string
		ViewTime time.Time
	)
	for _, db := range dbs {
		db1AuthorRows, err := db.Query("select w.web_name,v.uuid,vh.view_time from video_history vh inner join website w on w.id = vh.web_site_id inner join video v on v.id = vh.video_id where uuid != ''")
		if err != nil {
			fmt.Printf("sourceDataDb查询视频观看历史数据错误: %v\n", err)
			return
		}
		for db1AuthorRows.Next() {
			err = db1AuthorRows.Scan(&webName, &uuid, &ViewTime)
			if err != nil {
				fmt.Printf("sourceDataDb绑定数据错误: %v\n", err)
				continue
			}
			resultChan <- videoHistory{
				webName:  webName,
				uuid:     uuid,
				ViewTime: ViewTime,
			}
		}
		db1AuthorRows.Close()
	}
	close(resultChan)
}

func mergeData(sourceDataDb []*sql.DB, saveDb *sql.DB) {
	var (
		r               sql.Result
		err             error
		lastWebsiteName string
		lastWebsiteId   int64
		videoId         int64
		authorId        int64
	)
	videoInfoChan := make(chan videoInfo)
	go getVideoInfo(sourceDataDb, videoInfoChan)
	for dbVideoInfo := range videoInfoChan {
		// 查询website表中是否有该网站
		if lastWebsiteName != dbVideoInfo.webName {
			lastWebsiteName = dbVideoInfo.webName
			err = saveDb.QueryRow("select id from website where web_name = ?", dbVideoInfo.webName).Scan(&lastWebsiteId)
			if err != nil {
				if err == sql.ErrNoRows {
					r, err = saveDb.Exec("INSERT INTO website(web_name) values(?)", dbVideoInfo.webName)
					if err != nil {
						fmt.Printf("%s插入website失败: %s\n", dbVideoInfo.webName, err.Error())
						return
					}
					lastWebsiteId, _ = r.LastInsertId()
				} else {
					fmt.Printf("%s查询website失败: %s\n", dbVideoInfo.webName, err.Error())
					continue
				}
			}
		}
		// 查询author表中是否有该up
		err = saveDb.QueryRow("select id from author where author_web_uid = ?", dbVideoInfo.authorWebUid).Scan(&authorId)
		if err != nil {
			if err == sql.ErrNoRows {
				r, err = saveDb.Exec("INSERT INTO author(web_site_id, author_web_uid, author_name, author_desc, avatar) values(?, ?, ?, ?, ?)", lastWebsiteId, dbVideoInfo.authorWebUid, dbVideoInfo.authorName, dbVideoInfo.authorDesc, dbVideoInfo.avatar)
				if err != nil {
					fmt.Printf("%s插入author失败: %s\n", dbVideoInfo.authorName, err.Error())
					continue
				}
				authorId, _ = r.LastInsertId()
			} else {
				fmt.Printf("%s查询up失败: %s\n", dbVideoInfo.authorName, err.Error())
				continue
			}
		}
		// 查询video表中是否有该视频
		err = saveDb.QueryRow("select id from video where uuid = ?", dbVideoInfo.uuid).Scan(&videoId)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = saveDb.Exec("INSERT INTO video(web_site_id, author_id, title, cover_url, video_desc, duration, uuid, upload_time) values(?, ?, ?, ?, ?, ?, ?, ?)", lastWebsiteId, authorId, dbVideoInfo.videoTitle, dbVideoInfo.cover, dbVideoInfo.Desc, dbVideoInfo.duration, dbVideoInfo.uuid, dbVideoInfo.uploadTime)
				if err != nil {
					fmt.Printf("%s插入video失败: %s\n", dbVideoInfo.videoTitle, err.Error())
					continue
				}
			} else {
				fmt.Printf("%s查询video失败: %s\n", dbVideoInfo.uuid, err.Error())
				continue
			}
		}

	}

	videoHistoryChan := make(chan videoHistory)
	go getVideoHistory(sourceDataDb, videoHistoryChan)

	for dbVideoHistoryInfo := range videoHistoryChan {

		// 查询website表中是否有该网站
		if lastWebsiteName != dbVideoHistoryInfo.webName {
			lastWebsiteName = dbVideoHistoryInfo.webName
			err = saveDb.QueryRow("select id from website where web_name = ?", dbVideoHistoryInfo.webName).Scan(&lastWebsiteId)
			if err != nil {
				fmt.Printf("%s查询website失败: %s\n", dbVideoHistoryInfo.webName, err.Error())
				continue
			}
		}
		// 查询video表中是否有该视频
		err = saveDb.QueryRow("select id from video where uuid = ?", dbVideoHistoryInfo.uuid).Scan(&videoId)
		if err != nil {
			fmt.Printf("%s查询video失败: %s\n", dbVideoHistoryInfo.uuid, err.Error())
			continue
		}
		// 查询video_history表中是否有该视频
		err = saveDb.QueryRow("select id from video_history where video_id = ? and view_time = ?", videoId, dbVideoHistoryInfo.ViewTime).Scan(&videoId)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = saveDb.Exec("INSERT INTO video_history(web_site_id, video_id, view_time) values(?, ?, ?)", lastWebsiteId, videoId, dbVideoHistoryInfo.ViewTime)
				if err != nil {
					fmt.Printf("%s插入video_history失败: %s\n", dbVideoHistoryInfo.uuid, err.Error())
					continue
				}
			} else {
				fmt.Printf("%s查询video_history失败: %s\n", dbVideoHistoryInfo.uuid, err.Error())
				continue
			}
		}

	}
}
