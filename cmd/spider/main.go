package main

import (
	"database/sql"
	"fmt"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"path"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

var (
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
)

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList() []baseStruct.VideoInfo
}
type Spider struct {
	interval int64
}

func (s *Spider) getVideoInfo() {
	println("getVideoInfo")
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	for _, v := range videoCollection {
		website := v.GetWebSiteName()
		website.GetOrCreate(db)
		for _, video := range v.GetVideoList() {
			//fmt.Printf("video: %v\n", video)
			author := models.Author{AuthorName: video.AuthorName, WebSiteId: website.Id, AuthorWebUid: video.AuthorUuid}
			author.GetOrCreate(db)
			videoModel := models.Video{
				WebSiteId:  website.Id,
				AuthorId:   author.Id,
				Title:      video.Title,
				Desc:       video.Desc,
				Duration:   video.Duration,
				Url:        video.Url,
				Uuid:       video.VideoUuid,
				CoverUrl:   video.CoverUrl,
				UploadTime: video.PushTime,
				BiliOffset: video.Baseline,
			}
			videoModel.Save(db)
		}
	}
}

func (s *Spider) run(interface{}) {
	for _, v := range videoCollection {
		videoList := v.GetVideoList()
		for _, video := range videoList {
			fmt.Printf("video: %v\n", video)
		}
	}
	wheel.AppendOnceFunc(s.run, nil, "VideoSpider", timeWheel.Crontab{ExpiredTime: s.interval})
}

func main() {
	utils.InitLog(baseStruct.RootPath)
	models.InitDB(path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	videoCollection = []VideoCollection{
		bilibili.Bilibili,
	}
	wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
		IsRun: false,
		Log:   utils.TimeWheelLog,
	})
	wheel.AppendOnceFunc(spider.run, nil, "VideoSpider", timeWheel.Crontab{ExpiredTime: spider.interval})
	wheel.Start()
}
