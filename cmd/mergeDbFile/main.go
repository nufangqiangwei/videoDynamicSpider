package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strconv"
	"strings"
	"time"
)

// 将 E:\GoCode\videoDynamicAcquisition\videoInfo.db 文件中的video表
// 合并到 E:\GoCode\videoDynamicAcquisition\cmd\spider\videoInfo.db 文件中的video表
const saveDbPath = "E:\\GoCode\\videoDynamicAcquisition\\identifier.sqlite"

var sourceDbPath = []string{
	"E:\\GoCode\\videoDynamicAcquisition\\db\\videoInfo.db",
	"E:\\GoCode\\videoDynamicAcquisition\\db\\videoInfo111.db",
	"E:\\GoCode\\videoDynamicAcquisition\\db\\videoInfo222.db",
	"E:\\GoCode\\videoDynamicAcquisition\\db\\videoInfo333.db",
	"E:\\GoCode\\videoDynamicAcquisition\\db\\videoInfo444.db",
	"E:\\GoCode\\videoDynamicAcquisition\\videoInfo.db",
}
var (
	targetWebSiteMap       map[string]int64
	targetWebSiteAuthorMap map[string]map[string]int64
	resultData             map[string]videoInfo
)

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
	loadTargetDbData(saveDb)
	if targetWebSiteMap == nil || targetWebSiteAuthorMap == nil {
		return
	}
	videoInfoChan := make(chan videoInfo)
	go getVideoInfo([]*sql.DB{saveDb}, videoInfoChan)
	for dbVideoInfo := range videoInfoChan {
		resultData[dbVideoInfo.uuid] = dbVideoInfo
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
	mergeVideoInfo(sourceDBs, saveDb)
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
	webSiteId    int64
	authorId     int64
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
	)
	for _, db := range dbS {
		db1AuthorRows, err := db.Query("select w.web_name, a.author_web_uid, a.author_name, a.author_desc, a.avatar, v.title, v.cover_url, v.video_desc, v.duration, v.uuid from video v inner join author a on a.id = v.author_id inner join website w on w.id = v.web_site_id where a.author_web_uid!=0")
		if err != nil {
			fmt.Printf("sourceDataDb查询视频信息数据错误: %v\n", err)
			return
		}
		for db1AuthorRows.Next() {
			err = db1AuthorRows.Scan(&webName, &authorWebUid, &authorName, &authorDesc, &avatar, &videoTitle, &cover, &Desc, &duration, &uuid)
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
				_, err = saveDb.Exec("INSERT INTO video(web_site_id, author_id, title, cover_url, video_desc, duration, uuid, upload_time) values(?, ?, ?, ?, ?, ?, ?, ?)", lastWebsiteId, authorId, dbVideoInfo.videoTitle, dbVideoInfo.cover, dbVideoInfo.Desc, dbVideoInfo.duration, dbVideoInfo.uuid)
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

func mergeVideoInfo(sourceDataDb []*sql.DB, saveDb *sql.DB) {
	videoInfoChan := make(chan videoInfo)
	insertVideoData := make(map[string]videoInfo)
	insertAuthorData := make(map[string]videoInfo)
	updateData := make(map[string]videoInfo)
	go getVideoInfo(sourceDataDb, videoInfoChan)
	var (
		vi  videoInfo
		ok  bool
		r   sql.Result
		err error
	)
	for dbVideoInfo := range videoInfoChan {
		vi, ok = resultData[dbVideoInfo.uuid]
		if !ok {
			dbVideoInfo.webSiteId, ok = targetWebSiteMap[dbVideoInfo.webName]
			if !ok {
				r, err = saveDb.Exec("INSERT INTO website(web_name) values(?)", dbVideoInfo.webName)
				if err != nil {
					fmt.Printf("%s插入website失败: %s\n", dbVideoInfo.webName, err.Error())
					return
				}
				dbVideoInfo.webSiteId, _ = r.LastInsertId()
				targetWebSiteMap[dbVideoInfo.webName] = dbVideoInfo.webSiteId
			}
			dbVideoInfo.authorId, ok = targetWebSiteAuthorMap[dbVideoInfo.webName][dbVideoInfo.authorName]
			if !ok {
				r, err = saveDb.Exec("INSERT INTO author(web_site_id, author_web_uid, author_name, author_desc, avatar) values(?, ?, ?, ?, ?)", dbVideoInfo.webSiteId, dbVideoInfo.authorWebUid, dbVideoInfo.authorName, dbVideoInfo.authorDesc, dbVideoInfo.avatar)
				if err != nil {
					fmt.Printf("%s插入author失败: %s\n", dbVideoInfo.authorName, err.Error())
					continue
				}
				dbVideoInfo.authorId, _ = r.LastInsertId()
				if _, ok = targetWebSiteAuthorMap[dbVideoInfo.webName]; !ok {
					targetWebSiteAuthorMap[dbVideoInfo.webName] = make(map[string]int64)
				}
				targetWebSiteAuthorMap[dbVideoInfo.webName][dbVideoInfo.authorName] = dbVideoInfo.authorId
			}
			insertVideoData[dbVideoInfo.uuid] = dbVideoInfo
			insertAuthorData[dbVideoInfo.authorWebUid] = dbVideoInfo
		} else if (vi.Desc == "" && vi.Desc != dbVideoInfo.Desc) || (vi.videoTitle == "" && vi.videoTitle != dbVideoInfo.videoTitle) || (vi.cover == "" && vi.cover != dbVideoInfo.cover) {
			updateData[dbVideoInfo.uuid] = dbVideoInfo
		}
	}

	// 将insertAuthorData的数据写到数据库中，1000条一次
	// 将sql写入文件中
	file, _ := os.OpenFile("E:\\GoCode\\videoDynamicAcquisition\\cmd\\mergeDbFile\\insertAuthorData.sql", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()
	for _, sqlContext := range GetVideoInsertSQList(insertVideoData, 1000) {
		sqlContext := strings.Replace(sqlContext, "\n", "", -1)
		file.WriteString(sqlContext + ";\n")
		_, err := saveDb.Exec(sqlContext)
		if err != nil {
			fmt.Printf("插入video失败: %s\n", err.Error())
			continue
		}
	}

}

func mergeVideoHistory(sourceDataDb []*sql.DB, saveDb *sql.DB) {
	videoHistoryChan := make(chan videoHistory)
	go getVideoHistory(sourceDataDb, videoHistoryChan)
	for dbVideoHistoryInfo := range videoHistoryChan {
		// 查询video表中是否有该视频
		err := saveDb.QueryRow("select id from video where uuid = ?", dbVideoHistoryInfo.uuid).Scan(&dbVideoHistoryInfo.uuid)
		if err != nil {
			fmt.Printf("%s查询video失败: %s\n", dbVideoHistoryInfo.uuid, err.Error())
			continue
		}
		// 查询video_history表中是否有该视频
		err = saveDb.QueryRow("select id from video_history where video_id = ? and view_time = ?", dbVideoHistoryInfo.uuid, dbVideoHistoryInfo.ViewTime).Scan(&dbVideoHistoryInfo.uuid)
		if err != nil {
			if err == sql.ErrNoRows {
				_, err = saveDb.Exec("INSERT INTO video_history(web_site_id, video_id, view_time) values(?, ?, ?)", dbVideoHistoryInfo.webName, dbVideoHistoryInfo.uuid, dbVideoHistoryInfo.ViewTime)
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

func loadTargetDbData(saveDb *sql.DB) {
	r, err := saveDb.Query("select id, web_name from website")
	if err != nil {
		fmt.Printf("查询website失败: %s\n", err.Error())
		return
	}
	targetWebSiteMap = make(map[string]int64)
	targetWebSiteAuthorMap = make(map[string]map[string]int64)
	var (
		id           int64
		webName      string
		webSiteId    int64
		authorWebUid string
		authorName   string
		avatar       string
		authorDesc   string
		follow       bool
		followTime   time.Time
		createTime   time.Time
	)
	for r.Next() {
		err = r.Scan(&id, &webName)
		if err != nil {
			fmt.Printf("绑定website失败: %s\n", err.Error())
			continue
		}
		targetWebSiteMap[webName] = id
	}
	r.Close()
	//////////////////////////////////////////////
	// 数据结构 { WebSiteName :{authorName:models.Author}}
	r, err = saveDb.Query("select a.*,w.web_name from author a inner join website w on w.id = a.web_site_id")
	if err != nil {
		fmt.Printf("查询author失败: %s\n", err.Error())
		return
	}

	for r.Next() {
		err = r.Scan(&id, &webSiteId, &authorWebUid, &authorName, &avatar, &authorDesc, &follow, &followTime, &createTime, &webName)
		if err != nil {
			fmt.Printf("绑定author失败: %s\n", err.Error())
			continue
		}
		_, ok := targetWebSiteAuthorMap[webName]
		if !ok {
			targetWebSiteAuthorMap[webName] = make(map[string]int64)
		}
		targetWebSiteAuthorMap[webName][authorName] = id

	}
	r.Close()
}

const insertVideoHeader string = "INSERT INTO video(web_site_id, author_id, title, cover_url, video_desc, duration, uuid) VALUES"

func GetVideoInsertSQList(pointList map[string]videoInfo, groupSize int) []string {
	var sqlList []string
	sqlContext := ""
	i := 0
	for _, v := range pointList {
		if i%groupSize == 0 {
			if sqlContext != "" {
				//把上次拼接的SQL结果存储起来
				sqlList = append(sqlList, sqlContext)
			}
			//重置SQL
			sqlContext = insertVideoHeader
		}
		if sqlContext != insertVideoHeader {
			sqlContext = sqlContext + ","
		}
		sqlContext = fmt.Sprintf("%s(%d, %d, '%s', '%s', '%s', %d, '%s')",
			sqlContext,
			v.webSiteId,
			v.authorId,
			strconv.Quote(v.videoTitle),
			strconv.Quote(v.cover),
			strconv.Quote(v.Desc),
			v.duration,
			strconv.Quote(v.uuid),
		)
		i++
	}

	//把最后一次生成的SQL存储起来

	sqlList = append(sqlList, sqlContext)
	println(len(pointList))
	return sqlList
}
