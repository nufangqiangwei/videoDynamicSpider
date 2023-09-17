package main

import (
	"database/sql"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"path"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

const (
	defaultTicket = 60 * 5
	sixTime       = 3600 * 6
	twentyTime    = 3600 * 20
)

var (
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
)

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList(string) []baseStruct.VideoInfo
}
type Spider struct {
	interval int64
}

func getVideoInfo() {
	utils.Info.Println("getVideoInfo")
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	dynamicBaseLine := models.GetDynamicBaseline(db)
	for _, v := range videoCollection {
		website := v.GetWebSiteName()
		website.GetOrCreate(db)
		for index, video := range v.GetVideoList(dynamicBaseLine) {
			//fmt.Printf("video: %v\n", video)
			if index == 0 {
				models.SaveDynamicBaseline(db, video.Baseline)
			}
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
			}
			videoModel.Save(db)
		}
	}
}

func (s *Spider) getDynamic(interface{}) {
	getVideoInfo()
	s.interval = arrangeRunTime()
	_, err := wheel.AppendOnceFunc(s.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: s.interval})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

func main() {
	utils.InitLog(baseStruct.RootPath)
	models.InitDB(path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	videoCollection = []VideoCollection{
		bilibili.Spider,
	}
	wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
		IsRun: false,
		Log:   utils.TimeWheelLog,
	})
	spider = &Spider{
		interval: defaultTicket,
	}
	_, err := wheel.AppendOnceFunc(spider.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: spider.interval})
	if err != nil {
		return
	}
	wheel.Start()
}

func arrangeRunTime() int64 {
	nowTime := time.Now()
	zeroTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location())
	timeGap := nowTime.Sub(zeroTime) / 1000000000

	if sixTime > timeGap || timeGap > twentyTime {
		// 早上六点之前晚上八点之后，不再执行。六点之后才执行
		nextRunTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day()+1, 6, 0, 0, 0, nowTime.Location())
		return int64(nextRunTime.Sub(nowTime) / 1000000000)
	}
	return defaultTicket
}

// 定时抓取历史记录
func getHistory() {
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	baseLine := models.GetHistoryBaseLine(db)
	go bilibili.Spider.GetVideoHistoryList(baseLine)
	website := models.WebSite{WebName: "bilibili"}
	website.GetOrCreate(db)
	for {
		select {
		case videoInfo := <-bilibili.Spider.VideoHistoryChan:
			author := models.Author{AuthorName: videoInfo.AuthorName, WebSiteId: website.Id, AuthorWebUid: videoInfo.AuthorUuid}
			author.GetOrCreate(db)
			vi := models.Video{}
			vi.GetByUid(db, videoInfo.VideoUuid)
			if vi.Id <= 0 {
				vi.CreateTime = videoInfo.PushTime
				vi.Title = videoInfo.Title
				vi.Uuid = videoInfo.VideoUuid
				vi.AuthorId = author.Id
				vi.Save(db)
			}
			models.VideoHistory{
				WebSiteId: website.Id,
				VideoId:   vi.Id,
				ViewTime:  videoInfo.PushTime,
			}.Save(db)
		case baseLine = <-bilibili.Spider.VideoHistoryCloseChan:
			models.SaveHistoryBaseLine(db, baseLine)
			break
		}
	}
}
