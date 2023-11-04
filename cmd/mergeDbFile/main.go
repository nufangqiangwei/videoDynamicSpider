package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
	"time"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

// 将 E:\GoCode\videoDynamicAcquisition\videoInfo.db 文件中的video表
// 合并到 E:\GoCode\videoDynamicAcquisition\cmd\spider\videoInfo.db 文件中的video表
const saveDbPath = "E:\\GoCode\\videoDynamicAcquisition\\identifier.sqlite"

var sourceDbPath = []string{
	"C:\\Code\\GO\\videoDynamicSpider\\db\\videoInfo.db",
	"C:\\Code\\GO\\videoDynamicSpider\\db\\videoInfo111.db",
	"C:\\Code\\GO\\videoDynamicSpider\\db\\videoInfo222.db",
	"C:\\Code\\GO\\videoDynamicSpider\\db\\videoInfo333.db",
	"C:\\Code\\GO\\videoDynamicSpider\\db\\videoInfo444.db",
	"C:\\Code\\GO\\videoDynamicSpider\\videoInfo.db",
	"C:\\Code\\GO\\videoDynamicSpider\\videoInfo222.db",
}
var (
	targetWebSiteMap       map[string]int64
	targetWebSiteAuthorMap map[string]map[string]int64
	resultData             map[string]videoInfo
)

// 1. 打开两个数据库文件
// 2. 从第一个数据库中查询出所有的数据
// 3. 将第一个数据库中的数据插入到第二个数据库中
// 4. 关闭两个数据库文件
func main() {
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
		println(sourceDb)
		sourceDBs = append(sourceDBs, sourceDb)

	}
	mergeDB(sourceDBs)
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
	defer close(resultChan)
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
			continue
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

func saveVideoInfoToCsv(sourceDBs []*sql.DB) {
	videoInfoChan := make(chan videoInfo)
	go getVideoInfo(sourceDBs, videoInfoChan)
	file, _ := os.OpenFile("C:\\Code\\GO\\videoDynamicSpider\\source.csv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	file.WriteString("webName, authorWebUid, authorName, authorDesc, avatar, title, coverUrl, videoDesc, duration, uuid\n")
	var i int
	for dbVideoInfo := range videoInfoChan {
		i++
		s := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%d,%s", dbVideoInfo.webName, dbVideoInfo.authorWebUid, dbVideoInfo.authorName, dbVideoInfo.authorDesc, dbVideoInfo.avatar, dbVideoInfo.videoTitle, dbVideoInfo.cover, dbVideoInfo.Desc, dbVideoInfo.duration, dbVideoInfo.uuid)
		s = strings.Replace(s, "\n", "\\n", -1)
		s += "\n"
		_, err := file.WriteString(s)
		if err != nil {
			println(err.Error())
			return
		}
	}
	println(i)
	file.Close()
}

func mergeDB(sourceDBs []*sql.DB) {
	utils.InitLog("C:\\Code\\GO\\videoDynamicSpider")
	websiteMap := make(map[string]int64)
	var maxWebsiteId int64 = 0
	author := make(map[string]models.Author)
	var maxAuthorId int64 = 0
	video := make(map[string]models.Video)
	var maxVideoId int64 = 0
	videoInfoChan := make(chan videoInfo)
	go getVideoInfo(sourceDBs, videoInfoChan)
	for dbVideoInfo := range videoInfoChan {
		webId, ok := websiteMap[dbVideoInfo.webName]
		if !ok {
			maxWebsiteId++
			websiteMap[dbVideoInfo.webName] = maxWebsiteId
			webId = maxWebsiteId
		}
		authorModel, ok := author[dbVideoInfo.authorWebUid]
		if !ok {
			maxAuthorId++
			author[dbVideoInfo.authorWebUid] = models.Author{
				Id:           maxAuthorId,
				WebSiteId:    webId,
				AuthorWebUid: dbVideoInfo.authorWebUid,
				AuthorName:   dbVideoInfo.authorName,
				Avatar:       dbVideoInfo.avatar,
				Desc:         dbVideoInfo.authorDesc,
				Follow:       false,
				FollowTime:   time.Time{},
				Crawl:        false,
				CreateTime:   time.Time{},
			}
			authorModel = author[dbVideoInfo.authorWebUid]
		}
		_, ok = video[dbVideoInfo.uuid]
		if !ok {
			maxVideoId++
			video[dbVideoInfo.uuid] = models.Video{
				Id:         maxVideoId,
				WebSiteId:  webId,
				AuthorId:   authorModel.Id,
				Title:      dbVideoInfo.videoTitle,
				Desc:       dbVideoInfo.Desc,
				Duration:   dbVideoInfo.duration,
				Uuid:       dbVideoInfo.uuid,
				Url:        "",
				CoverUrl:   dbVideoInfo.cover,
				UploadTime: time.Time{},
				CreateTime: time.Time{},
			}
		}
	}

	saveDb, _ := sql.Open("sqlite3", "C:\\Code\\GO\\videoDynamicSpider\\newDb.db")
	baseModels := []models.BaseModel{
		&models.WebSite{},
		&models.Author{},
		&models.Video{},
		&models.BiliAuthorVideoNumber{},
		&models.BiliSpiderHistory{},
		&models.VideoHistory{},
		&models.Collect{},
		&models.CollectVideo{},
	}
	for _, baseModel := range baseModels {
		_, err := saveDb.Exec(baseModel.CreateTale())
		if err != nil {
			utils.ErrorLog.Println("创建表失败")
			utils.ErrorLog.Println(err.Error())
			utils.ErrorLog.Println(baseModel.CreateTale())
		}
	}
	//for _, authorM := range author {
	//	saveDb.Exec("INSERT INTO author (id,web_site_id, author_web_uid,author_name,avatar,author_desc,follow,follow_time,crawl) VALUES (?,?,?,?,?,?,?,?,?)", authorM.Id, authorM.WebSiteId, authorM.AuthorWebUid, authorM.AuthorName, authorM.Avatar, authorM.Desc, authorM.Follow, authorM.FollowTime, authorM.Crawl)
	//}
	//for _, videoM := range video {
	//	saveDb.Exec("INSERT INTO video (id,web_site_id, author_id, title,video_desc,duration,uuid, url, cover_url) VALUES (?,?, ?, ?, ?,?,?,?,?)",
	//		videoM.Id, videoM.WebSiteId, videoM.AuthorId, videoM.Title, videoM.Desc, videoM.Duration, videoM.Uuid, videoM.Url, videoM.CoverUrl)
	//}
	resultChan := make(chan videoHistory)
	videoHistoryMap := make(map[string][]time.Time)
	go getVideoHistory(sourceDBs, resultChan)
	var have bool
	for videoHistoryData := range resultChan {
		ti, ok := videoHistoryMap[videoHistoryData.uuid]
		if !ok {
			videoHistoryMap[videoHistoryData.uuid] = []time.Time{
				videoHistoryData.ViewTime,
			}
			continue
		}
		have = false
		for _, i := range ti {
			if i == videoHistoryData.ViewTime {
				have = true
			}
		}
		if !have {
			videoHistoryMap[videoHistoryData.uuid] = append(ti, videoHistoryData.ViewTime)
		}
	}

	for uuid, ti := range videoHistoryMap {
		videoId := video[uuid].Id
		insertData := make([]string, 0)
		for _, i := range ti {
			insertData = append(insertData, fmt.Sprintf("(%d,%d,'%s')", 1, videoId, i.Format("2006.01.02 15:04:05")))
		}
		saveDb.Exec(fmt.Sprintf("insert into video_history (web_site_id, `video_id`,view_time ) values %s ;", strings.Join(insertData, ",")))
	}
}
