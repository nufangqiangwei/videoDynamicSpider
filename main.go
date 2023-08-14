package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nufangqiangwei/timewheel"
	"path"
	"time"
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
	Uuid       string
	Url        string
	CoverUrl   string
	AuthorName string
	AuthorUrl  string
	PushTime   time.Time
}

const sqliteDaName = "videoInfo.db"

type VideoCollection interface {
	getWebSiteName() models.WebSite
	getVideoList() []VideoInfo
}

type Spider struct {
	interval int64
}

func (s *Spider) getVideoInfo() {
	db, _ := sql.Open("sqlite3", path.Join("E:\\GoCode\\videoDynamicAcquisition", sqliteDaName))
	for _, v := range videoCollection {
		website := v.getWebSiteName()
		website.GetOrCreate(db)
		videoList := v.getVideoList()
		for _, video := range videoList {
			fmt.Printf("video: %v\n", video)
			author := models.Author{AuthorName: video.AuthorName}
			author.GetOrCreate(db)
			videoModel := models.Video{
				WebSiteId:  website.Id,
				AuthorId:   author.Id,
				Title:      video.Title,
				Url:        video.Url,
				CoverUrl:   video.CoverUrl,
				CreateTime: video.PushTime,
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

func main() {
	println("hello world")
	InitLog("E:\\GoCode\\videoDynamicAcquisition")
	models.InitDB(path.Join("E:\\GoCode\\videoDynamicAcquisition", sqliteDaName))
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
	println("end world")
}
