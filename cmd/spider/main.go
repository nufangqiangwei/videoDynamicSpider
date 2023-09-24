package main

import (
	"database/sql"
	"fmt"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"path"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

const (
	defaultTicket = 60 * 5
	historyTicket = 60 * 60
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

func (s *Spider) getDynamic(interface{}) {
	utils.Info.Println("getVideoInfo")
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	dynamicBaseLine := models.GetDynamicBaseline(db)
	for _, v := range videoCollection {
		website := v.GetWebSiteName()
		website.GetOrCreate(db)
		for index, video := range v.GetVideoList(dynamicBaseLine) {
			fmt.Printf("video: %s\n", video.Title)
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
	db.Close()
	_, err := wheel.AppendOnceFunc(s.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(defaultTicket)})
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
	_, err := wheel.AppendOnceFunc(spider.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: defaultTicket})
	if err != nil {
		return
	}
	_, err = wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: historyTicket})
	if err != nil {
		return
	}
	wheel.Start()
}

func arrangeRunTime(defaultValue int64) int64 {
	nowTime := time.Now()
	zeroTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location())
	timeGap := nowTime.Sub(zeroTime) / 1000000000

	if sixTime > timeGap || timeGap > twentyTime {
		// 早上六点之前晚上八点之后，不再执行。六点之后才执行
		nextRunTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day()+1, 6, 0, 0, 0, nowTime.Location())
		return int64(nextRunTime.Sub(nowTime) / 1000000000)
	}
	return defaultValue
}

// 定时抓取历史记录 历史数据，同步到以观看表,视频信息储存到视频表，作者信息储存到作者表。新数据标记未同步历史数据，未关注
func (s *Spider) getHistory(interface{}) {
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	baseLine := models.GetHistoryBaseLine(db)
	var (
		lastHistoryTimestamp int64 = 0
		err                  error
	)
	if baseLine != "" {
		lastHistoryTimestamp, err = strconv.ParseInt(baseLine, 10, 64)
		if err != nil {
			utils.ErrorLog.Println("获取历史基线失败")
			return
		}
	}
	go bilibili.Spider.GetVideoHistoryList(lastHistoryTimestamp)
	website := models.WebSite{WebName: "bilibili"}
	website.GetOrCreate(db)
	for {
		select {
		case videoInfo := <-bilibili.Spider.VideoHistoryChan:
			author := models.Author{
				AuthorName:   videoInfo.AuthorName,
				WebSiteId:    website.Id,
				AuthorWebUid: videoInfo.AuthorUuid,
				Crawl:        true,
			}
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
		case newestTimestamp := <-bilibili.Spider.VideoHistoryCloseChan:
			models.SaveHistoryBaseLine(db, strconv.FormatInt(newestTimestamp, 10))
			if wheel != nil {
				_, err := wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: historyTicket})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
					return
				}
			}
			return
		}
	}
}

// 更新收藏夹列表，喝订阅的合集列表，新创建的同步视频数据
func (s *Spider) updateCollectList(interface{}) {
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	web := models.WebSite{
		WebName: "bilibili",
	}
	web.GetOrCreate(db)
	newCollectList := bilibili.Spider.GetCollectList(db)
	for _, collectId := range newCollectList.Collect {
		for _, info := range bilibili.Spider.GetCollectAllVideo(collectId, 0) {
			author := models.Author{
				WebSiteId:    web.Id,
				AuthorWebUid: strconv.Itoa(info.Upper.Mid),
				AuthorName:   info.Upper.Name,
				Avatar:       info.Upper.Face,
				Follow:       false,
				Crawl:        true,
			}
			author.GetOrCreate(db)
			vi := models.Video{}
			vi.GetByUid(db, info.Bvid)
			if vi.Id <= 0 {
				vi.WebSiteId = web.Id
				vi.AuthorId = author.Id
				vi.Title = info.Title
				vi.Desc = info.Intro
				vi.Duration = info.Duration
				vi.Uuid = info.Bvid
				vi.CoverUrl = info.Cover
				vi.UploadTime = time.Unix(info.Ctime, 0)
				vi.Save(db)
			}
			models.CollectVideo{
				CollectId: collectId,
				VideoId:   vi.Id,
				Mtime:     time.Unix(info.FavTime, 0),
			}.Save(db)
		}
	}
	for _, collectId := range newCollectList.Season {
		for _, info := range bilibili.Spider.GetSeasonAllVideo(collectId) {
			author := models.Author{
				WebSiteId:    web.Id,
				AuthorWebUid: strconv.Itoa(info.Upper.Mid),
				AuthorName:   info.Upper.Name,
				Follow:       false,
				Crawl:        true,
			}
			author.GetOrCreate(db)
			vi := models.Video{}
			vi.GetByUid(db, info.Bvid)
			if vi.Id <= 0 {
				vi.WebSiteId = web.Id
				vi.AuthorId = author.Id
				vi.Title = info.Title
				vi.Duration = info.Duration
				vi.Uuid = info.Bvid
				vi.CoverUrl = info.Cover
				vi.Save(db)
			}
			models.CollectVideo{
				CollectId: collectId,
				VideoId:   vi.Id,
			}.Save(db)

		}
		time.Sleep(time.Second * 5)
	}
}

// 更新收藏夹视频列表，更新合集视频列表
func updateCollectVideoList(interface{}) {
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	query, err := db.Query("select cv.collect_id,v.uuid from collect_video cv inner join collect c on c.bv_id = cv.collect_id inner join video v on v.id = cv.video_id where c.`type` = 1 order by cv.collect_id,mtime desc")
	queryResult := make(map[int64][]string)
	if err != nil {
		utils.ErrorLog.Printf("查询收藏夹视频数量失败：%s\n", err.Error())
		return
	}
	var (
		collectId int64
		count     string
		countList []string
	)
	for query.Next() {
		err = query.Scan(&collectId, &count)
		if err != nil {
			continue
		}
		_, ok := queryResult[collectId]
		if !ok {
			queryResult[collectId] = []string{}
		}
		queryResult[collectId] = append(queryResult[collectId], count)
	}
	query.Close()
	waitUpdateList := make([]int64, 0)
	for collectId, countList = range queryResult {
		r := bilibili.GetCollectVideoList(collectId)
		// countList 和r.Data中的BvId对比，找不同的值
		for _, info := range r.Data {
			if !utils.InArray(info.BvId, countList) {
				waitUpdateList = append(waitUpdateList, collectId)
				break
			}
		}
	}
	// 没有待更新的内容
	if len(waitUpdateList) == 0 {
		return
	}
	web := models.WebSite{
		WebName: "bilibili",
	}
	web.GetOrCreate(db)
	for _, collectId = range waitUpdateList {
		r := bilibili.Spider.GetCollectAllVideo(collectId, 1)
		for _, info := range r {
			if !utils.InArray(info.BvId, countList) {
				author := models.Author{
					WebSiteId:    web.Id,
					AuthorWebUid: strconv.Itoa(info.Upper.Mid),
					AuthorName:   info.Upper.Name,
					Follow:       false,
					Crawl:        true,
				}
				author.GetOrCreate(db)
				vi := models.Video{}
				vi.GetByUid(db, info.Bvid)
				if vi.Id <= 0 {
					vi.WebSiteId = web.Id
					vi.AuthorId = author.Id
					vi.Title = info.Title
					vi.Duration = info.Duration
					vi.Uuid = info.Bvid
					vi.CoverUrl = info.Cover
					vi.Save(db)
				}
				models.CollectVideo{
					CollectId: collectId,
					VideoId:   vi.Id,
				}.Save(db)
			}
		}
	}
}

// 同步关注信息
func (s *Spider) updateFollowInfo(interface{}) {
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	web := models.WebSite{
		WebName: "bilibili",
	}
	web.GetOrCreate(db)
	upList := bilibili.Spider.GetFollowingList()
	r, err := db.Query("select author_web_uid from main.author where follow=1")
	if err != nil {
		utils.ErrorLog.Printf("查询当前关注失败：%s\n", err.Error())
		return
	}
	var (
		followList       []string
		authorWebUid     string
		nowFollowList    []string
		notFollowList    []string
		appendFollowList []string
	)
	for r.Next() {
		err = r.Scan(&authorWebUid)
		if err != nil {
			continue
		}
		followList = append(followList, authorWebUid)
	}
	for _, up := range upList {
		nowFollowList = append(nowFollowList, strconv.FormatInt(up.Mid, 10))
	}
	notFollowList = utils.ArrayDifference(followList, nowFollowList)
	appendFollowList = utils.ArrayDifference(nowFollowList, followList)
	db.Exec("update main.author set main.author.follow=0 where main.author.author_web_uid in ?", notFollowList)
	for _, upInfo := range upList {
		if utils.InArray(strconv.FormatInt(upInfo.Mid, 10), appendFollowList) {
			author := models.Author{
				WebSiteId:    web.Id,
				AuthorWebUid: strconv.FormatInt(upInfo.Mid, 10),
				AuthorName:   upInfo.Uname,
				Avatar:       upInfo.Face,
				Desc:         upInfo.Sign,
				Follow:       true,
				FollowTime:   time.Unix(upInfo.Mtime, 0),
			}
			author.UpdateOrCreate(db)
		}
	}

}
