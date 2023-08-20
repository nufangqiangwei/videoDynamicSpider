package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nufangqiangwei/timewheel"
	"path"
	"strconv"
	"videoDynamicAcquisition/models"
)

var (
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
)

type VideoInfo struct {
	WebSite    string
	Title      string
	VideoUuid  string
	Url        string
	CoverUrl   string
	AuthorUuid string
	AuthorName string
	AuthorUrl  string
	PushTime   int64
}

const sqliteDaName = "videoInfo.db"
const rootPath = "C:\\Code\\GO\\videoDynamicSpider"

type VideoCollection interface {
	getWebSiteName() models.WebSite
	getVideoList() []VideoInfo
}

type Spider struct {
	interval int64
}

func (s *Spider) getVideoInfo() {
	db, _ := sql.Open("sqlite3", path.Join(rootPath, sqliteDaName))
	for _, v := range videoCollection {
		website := v.getWebSiteName()
		website.GetOrCreate(db)
		for _, video := range v.getVideoList() {
			//fmt.Printf("video: %v\n", video)
			author := models.Author{AuthorName: video.AuthorName, WebSiteId: website.Id, AuthorWebUid: video.AuthorUuid}
			author.GetOrCreate(db)
			videoModel := models.Video{
				WebSiteId:  website.Id,
				AuthorId:   author.Id,
				Title:      video.Title,
				Url:        video.Url,
				Uuid:       video.VideoUuid,
				CoverUrl:   video.CoverUrl,
				UploadTime: video.PushTime,
			}
			videoModel.Save(db)
		}
	}
}

func (s *Spider) run(interface{}) {
	for _, v := range videoCollection {
		videoList := v.getVideoList()
		for _, video := range videoList {
			fmt.Printf("video: %v\n", video)
		}
	}
	wheel.AppendOnceFunc(s.run, nil, "VideoSpider", timeWheel.Crontab{ExpiredTime: s.interval})
}

func getVideoList(ctx *gin.Context) {
	db, _ := sql.Open("sqlite3", path.Join(rootPath, sqliteDaName))
	webSiteName := ctx.DefaultQuery("webSite", "bilibili")
	pageSTR := ctx.DefaultQuery("page", "1")
	sizeSTR := ctx.DefaultQuery("size", "30")
	page, err := strconv.ParseInt(pageSTR, 10, 32)
	if err != nil {
		page = 1
	}
	size, err := strconv.ParseInt(sizeSTR, 10, 32)
	if err != nil {
		size = 30
	}
	rows, err := db.Query("select v.uuid, v.cover_url, v.url, a.author_name, a.author_web_uid,v.upload_time from video v left join author a on v.author_id = a.id where v.web_site_id = (select id from website where web_name = '%s') order by v.upload_time desc limit %d,%d;", webSiteName, (page-1)*size, page*size)
	result := map[string][]VideoInfo{"data": []VideoInfo{}}
	if err != nil {
		ctx.JSONP(200, result)
		return
	}
	for rows.Next() {
		rows.Scan()
	}
}

func updateCookies(ctx *gin.Context) {

}

func main() {
	//InitLog("E:\\GoCode\\videoDynamicAcquisition")
	models.InitDB(path.Join("C:\\Code\\GO\\videoDynamicSpider", sqliteDaName))
	videoCollection = []VideoCollection{
		makeBilibiliSpider(),
	}
	spider = &Spider{interval: 60 * 5}
	//wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
	//	IsRun: true,
	//	Log:   timewheelLog,
	//})
	//wheel.AppendOnceFunc(spider.run, nil, "VideoSpider", timeWheel.Crontab{ExpiredTime: spider.interval})
	spider.getVideoInfo()
	server := gin.Default()
	server.GET("/getVideoList", getVideoList)
	server.GET("/updateCookies", updateCookies)
	server.Run("localhost:8000")
}
