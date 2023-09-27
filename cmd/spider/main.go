package main

import (
	timeWheel "github.com/nufangqiangwei/timewheel"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

// 数据库下载地址： https://101.32.15.231/icon/videoInfo.db
const (
	defaultTicket = 60 * 5
	oneTicket     = 60 * 60
	sixTime       = oneTicket * 6
	twentyTime    = oneTicket * 20
	twelveTicket  = oneTicket * 12
)

var (
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
)

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList(string, chan<- baseStruct.VideoInfo, chan<- baseStruct.TaskClose)
}
type Spider struct {
	interval int64
}

func init() {
	utils.InitLog(baseStruct.RootPath)
	baseStruct.InitDB()
	wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
		IsRun: false,
		Log:   utils.TimeWheelLog,
	})
}

func main() {
	videoCollection = []VideoCollection{
		bilibili.Spider,
	}
	spider = &Spider{
		interval: defaultTicket,
	}
	_, err := wheel.AppendOnceFunc(spider.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: defaultTicket})
	if err != nil {
		return
	}
	_, err = wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: oneTicket + 523})
	if err != nil {
		return
	}
	_, err = wheel.AppendOnceFunc(spider.updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + 120})
	if err != nil {
		return
	}
	_, err = wheel.AppendOnceFunc(spider.updateCollectVideoList, nil, "updateCollectVideoSpider", timeWheel.Crontab{ExpiredTime: oneTicket})
	if err != nil {
		return
	}
	_, err = wheel.AppendOnceFunc(spider.updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
	if err != nil {
		return
	}
	wheel.Start()
}

func arrangeRunTime(defaultValue, leftBorder, rightBorder int64) int64 {
	nowTime := time.Now()
	zeroTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location())
	timeGap := int64(nowTime.Sub(zeroTime) / 1000000000)

	if leftBorder > timeGap || timeGap > rightBorder {
		// 早上六点之前晚上八点之后，不再执行。六点之后才执行
		nextRunTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day()+1, 6, 0, 0, 0, nowTime.Location())
		return int64(nextRunTime.Sub(nowTime) / 1000000000)
	}
	return defaultValue
}

func (s *Spider) getDynamic(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(s.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(defaultTicket, sixTime, twentyTime)})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	utils.Info.Println("getVideoInfo")
	db := baseStruct.CanUserDb()
	defer db.Close()
	dynamicBaseLine := models.GetDynamicBaseline(db)
	videoResultChan := make(chan baseStruct.VideoInfo)
	closeChan := make(chan baseStruct.TaskClose)
	runWebSite := make([]string, 0)

	for _, v := range videoCollection {
		go v.GetVideoList(dynamicBaseLine, videoResultChan, closeChan)
		runWebSite = append(runWebSite, v.GetWebSiteName().WebName)
	}

	var (
		videoInfo baseStruct.VideoInfo
		closeInfo baseStruct.TaskClose
	)
	for {
		select {
		case videoInfo = <-videoResultChan:
			website := models.WebSite{WebName: videoInfo.WebSite}
			website.GetOrCreate(db)
			author := models.Author{AuthorName: videoInfo.AuthorName, WebSiteId: website.Id, AuthorWebUid: videoInfo.AuthorUuid}
			author.GetOrCreate(db)
			videoModel := models.Video{
				WebSiteId:  website.Id,
				AuthorId:   author.Id,
				Title:      videoInfo.Title,
				Desc:       videoInfo.Desc,
				Duration:   videoInfo.Duration,
				Url:        videoInfo.Url,
				Uuid:       videoInfo.VideoUuid,
				CoverUrl:   videoInfo.CoverUrl,
				UploadTime: videoInfo.PushTime,
			}
			videoModel.Save(db)
		case closeInfo = <-closeChan:
			// 删除closeInfo.WebSite的任务
			for index, v := range runWebSite {
				if v == closeInfo.WebSite {
					runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
					break
				}
			}
			if closeInfo.WebSite == "bilibili" && closeInfo.Code > 0 {
				models.SaveDynamicBaseline(db, strconv.Itoa(closeInfo.Code))
			}
		}
		if len(runWebSite) == 0 {
			break
		}
	}
	close(closeChan)
	close(videoResultChan)

	_, err := wheel.AppendOnceFunc(s.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(defaultTicket, sixTime, twentyTime)})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

// 定时抓取历史记录 历史数据，同步到以观看表,视频信息储存到视频表，作者信息储存到作者表。新数据标记未同步历史数据，未关注
func (s *Spider) getHistory(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: oneTicket})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	db := baseStruct.CanUserDb()
	defer db.Close()
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
	VideoHistoryChan := make(chan baseStruct.VideoInfo)
	VideoHistoryCloseChan := make(chan int64)
	go bilibili.Spider.GetVideoHistoryList(lastHistoryTimestamp, VideoHistoryChan, VideoHistoryCloseChan)
	website := models.WebSite{WebName: "bilibili"}
	website.GetOrCreate(db)
	for {
		select {
		case videoInfo := <-VideoHistoryChan:
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
		case newestTimestamp := <-VideoHistoryCloseChan:
			models.SaveHistoryBaseLine(db, strconv.FormatInt(newestTimestamp, 10))
			_, err := wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: oneTicket})
			if err != nil {
				utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				return
			}
			return
		}
	}
}

// 更新收藏夹列表，喝订阅的合集列表，新创建的同步视频数据
func (s *Spider) updateCollectList(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(spider.updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	db := baseStruct.CanUserDb()
	defer db.Close()
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

	_, err := wheel.AppendOnceFunc(spider.updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

// 更新收藏夹视频列表，更新合集视频列表
func (s *Spider) updateCollectVideoList(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(spider.updateCollectVideoList, nil, "updateCollectVideoSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(twelveTicket, sixTime, twentyTime)})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	db := baseStruct.CanUserDb()
	defer db.Close()

	query, err := db.Query("select cv.collect_id,v.uuid from collect_video cv inner join collect c on c.bv_id = cv.collect_id inner join video v on v.id = cv.video_id where c.`type` = 1 and mtime>'0001-01-01 00:00:00+00:00' order by cv.collect_id,mtime desc")
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
		countList = queryResult[collectId]
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

	_, err = wheel.AppendOnceFunc(spider.updateCollectVideoList, nil, "updateCollectVideoSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(twelveTicket, sixTime, twentyTime)})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

// 同步关注信息
func (s *Spider) updateFollowInfo(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(spider.updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	db := baseStruct.CanUserDb()
	defer db.Close()
	web := models.WebSite{
		WebName: "bilibili",
	}
	web.GetOrCreate(db)
	resultChan := make(chan bilibili.FollowingUP)
	closeChan := make(chan int64)
	go bilibili.Spider.GetFollowingList(resultChan, closeChan)
	r, err := db.Query("select author_web_uid from main.author where follow=1")
	if err != nil {
		utils.ErrorLog.Printf("查询当前关注失败：%s\n", err.Error())
		return
	}
	var (
		followList   []string
		authorWebUid string
		upInfo       bilibili.FollowingUP
		close        bool
		//nowFollowList []string
		//notFollowList    []string
		//appendFollowList []string
	)
	for r.Next() {
		err = r.Scan(&authorWebUid)
		if err != nil {
			continue
		}
		followList = append(followList, authorWebUid)
	}
	//for _, up := range upList {
	//	nowFollowList = append(nowFollowList, strconv.FormatInt(up.Mid, 10))
	//}
	//notFollowList = utils.ArrayDifference(followList, nowFollowList)
	//appendFollowList = utils.ArrayDifference(nowFollowList, followList)
	//db.Exec("update main.author set main.author.follow=0 where main.author.author_web_uid in ?", notFollowList)
	//if utils.InArray(strconv.FormatInt(upInfo.Mid, 10), followList) {
	//}
	for {
		select {
		case upInfo = <-resultChan:
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
		case <-closeChan:
			close = true
		}
		if close {
			break
		}
	}

	_, err = wheel.AppendOnceFunc(spider.updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

func deleteRepeatAuthor() {
	db := baseStruct.CanUserDb()
	defer db.Close()
	//select author_name
	//from author
	//group by author_name
	//having count(*) > 1;

	repeatAuthor, err := db.Query("select author_name from author group by author_name having count(*) > 1")
	if err != nil {
		return
	}
	var (
		authorName string
		authorList []string
	)
	for repeatAuthor.Next() {
		err = repeatAuthor.Scan(&authorName)
		if err != nil {
			continue
		}
		authorList = append(authorList, authorName)
	}
	repeatAuthor.Close()
	for _, authorName = range authorList {
		_, err = db.Exec("update video set author_id  = (select id from author where author_name = ? and author.web_site_id = 1),    web_site_id=1 where author_id = (select id from author where author_name = ? and author.web_site_id = 0);", authorName, authorName)
		if err != nil {
			println(err.Error())
			continue
		}
	}

}
