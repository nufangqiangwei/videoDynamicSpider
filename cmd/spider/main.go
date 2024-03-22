package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"math/rand"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

const (
	oneMinute     = 60
	defaultTicket = 60 * 5
	oneTicket     = 60 * 60
	sixTime       = oneTicket * 6
	twentyTime    = oneTicket * 20
	twelveTicket  = oneTicket * 12
	configPath    = "./config.json"
)

var (
	videoCollection []models.VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
	config          *utils.Config
	wheelLog        log.LogInputFile
	databaseLog     log.LogInputFile
)

type Spider struct {
	interval int64
}

func readConfig() error {
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		println(err.Error())
		return err
	}
	config = &utils.Config{}
	err = json.Unmarshal(fileData, config)
	if err != nil {
		println(err.Error())
		return err
	}
	fmt.Printf("%v\n", *config)
	return nil
}

func init() {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		println("时区设置错误")
		os.Exit(2)
		return
	}
	time.Local = location

	// 读取配置文件
	err = readConfig()
	if err != nil {
		os.Exit(2)
		return
	}
	if config.DataPath != "" {
		baseStruct.RootPath = config.DataPath
	}

	// 初始化日志
	logBlockList := log.InitLog(path.Join(baseStruct.RootPath, "log"), "TimeWheel", "database")
	for _, logBlock := range logBlockList {
		if logBlock.FileName == "TimeWheel" {
			wheelLog = logBlock
			continue
		}
		if logBlock.FileName == "databaseLog" {
			databaseLog = logBlock
			continue
		}
	}

	// 初始化数据库，连接到数据库
	models.InitDB(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User, config.DB.Password, config.DB.HOST, config.DB.Port, config.DB.DatabaseName), false, databaseLog.WriterObject)

	// 初始化加载cookies模块
	cookies.DataSource = models.WebSiteCookies{}
	cookies.FlushAllCookies()
	initWebSiteSpider()

	// 初始化定时器
	wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
		IsRun: false,
		Log:   wheelLog.WriterObject,
	})
	rand.Seed(time.Now().UnixNano())
	log.Info.Println("初始化完成：", time.Now().Format("2006.01.02 15:04:05"))
}

func initWebSiteSpider() {
	w := models.WebSite{}
	models.GormDB.Where("web_name=?", bilibili.Spider.GetWebSiteName().WebName).First(&w)
	if w.Id == 0 {
		panic(fmt.Sprintf("%s站点在数据库中找不到对应的数据。", bilibili.Spider.GetWebSiteName().WebName))
	}
	var (
		dynamicBaseLine map[string]int
		historyBaseLine map[string]int64
	)
	dynamicBaseLine = make(map[string]int)
	historyBaseLine = make(map[string]int64)
	cookies.RangeCookiesMap(func(webSiteName, userName string, userCookie *cookies.UserCookie) {
		userId, err := models.GetAuthorId(userName)
		if err != nil {
			log.ErrorLog.Printf("获取%s用户id失败：%s\n", userName, err.Error())
			return
		}
		userCookie.SetDBPrimaryKeyId(userId)

		if webSiteName == bilibili.Spider.GetWebSiteName().WebName {
			var (
				result               []models.BiliSpiderHistory
				err                  error
				intLatestBaseline    int
				lastHistoryTimestamp int64
			)
			models.GormDB.Model(&models.BiliSpiderHistory{}).Where("author_id = ?", userId).Find(&result)
			for _, rowData := range result {
				if rowData.KeyName == "dynamic_baseline" {
					intLatestBaseline, err = strconv.Atoi(rowData.Values)
					if err != nil {
						log.ErrorLog.Printf("转换%s的dynamic_baseline失败：%s\n", userName, err.Error())
						continue
					}
					dynamicBaseLine[userName] = intLatestBaseline
				}
				if rowData.KeyName == "history_baseline" {
					lastHistoryTimestamp, err = strconv.ParseInt(rowData.Values, 10, 64)
					if err != nil {
						log.ErrorLog.Printf("转换%s的history_baseline失败：%s\n", userName, err.Error())
						continue
					}
					historyBaseLine[userName] = lastHistoryTimestamp
				}
			}
		}
	})
	bilibili.Spider.Init(dynamicBaseLine, historyBaseLine, w.Id)
}

func main() {
	videoCollection = []models.VideoCollection{
		bilibili.Spider,
	}
	spider = &Spider{
		interval: defaultTicket,
	}
	var err error
	_, err = wheel.AppendOnceFunc(spider.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: defaultTicket})
	if err != nil {
		return
	}
	_, err = wheel.AppendOnceFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: 10})
	if err != nil {
		return
	}
	//_, err = wheel.AppendOnceFunc(updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + 120})
	//if err != nil {
	//	return
	//}
	_, err = wheel.AppendOnceFunc(updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
	if err != nil {
		return
	}
	//_, err = wheel.AppendCycleFunc(runToDoTask, nil, "pushTaskToProxy", timeWheel.Crontab{ExpiredTime: oneMinute})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendCycleFunc(checkProxyTaskStatus, nil, "getTaskStatus", timeWheel.Crontab{ExpiredTime: oneMinute})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendCycleFunc(loadProxyInfo, nil, "loadConfigProxyInfo", timeWheel.Crontab{ExpiredTime: defaultTicket})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendCycleFunc(downloadProxyTaskDataFile, nil, "downloadProxyTaskDataFile", timeWheel.Crontab{ExpiredTime: oneTicket})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendOnceFunc(readPath, nil, "importProxyFileData", timeWheel.Crontab{ExpiredTime: 60})
	//if err != nil {
	//	return
	//}
	wheel.Start()
}

func arrangeRunTime(defaultValue, leftBorder, rightBorder int64) int64 {
	nowTime := time.Now()
	zeroTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location())
	timeGap := int64(nowTime.Sub(zeroTime) / 1000000000)

	// 早上六点之前晚上八点之后，不再执行。六点之后才执行
	if leftBorder > timeGap || timeGap > rightBorder {
		nextRunTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day()+1, 6, 0, 0, 0, nowTime.Location())
		return int64(nextRunTime.Sub(nowTime)/1000000000) + rand.Int63n(100)
	}
	return defaultValue + rand.Int63n(100)
}

func (s *Spider) getDynamic(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(s.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(defaultTicket, sixTime, twentyTime)})
				if err != nil {
					log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	videoResultChan := make(chan models.Video)
	closeChan := make(chan models.TaskClose)
	runWebSite := make([]string, 0)

	for _, v := range videoCollection {
		go v.GetVideoList(videoResultChan, closeChan)
		runWebSite = append(runWebSite, v.GetWebSiteName().WebName)
	}

	var (
		videoInfo models.Video
		closeInfo models.TaskClose
		err       error
	)
	for {
		select {
		case videoInfo = <-videoResultChan:
			videoInfo.UpdateVideo()
		case closeInfo = <-closeChan:
			// 删除closeInfo.WebSite的任务
			for index, v := range runWebSite {
				if v == closeInfo.WebSite {
					runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
					break
				}
			}
			if closeInfo.WebSite == bilibili.Spider.GetWebSiteName().WebName {
				for _, info := range closeInfo.Data {
					err = models.SaveSpiderParamByUserId(info.UserId, "dynamic_baseline", info.EndBaseLine)
					if err != nil {
						log.ErrorLog.Printf("保存dynamic_baseline失败：%s\n", err.Error())
					}
				}
			}
		}
		if len(runWebSite) == 0 {
			break
		}
	}
	close(closeChan)
	close(videoResultChan)
	nextRunTime := arrangeRunTime(defaultTicket, sixTime, twentyTime)
	log.Info.Printf("%d秒后再次抓取动态", nextRunTime)
	_, err = wheel.AppendOnceFunc(s.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: nextRunTime})
	if err != nil {
		log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

// 定时抓取历史记录 历史数据，同步到以观看表,视频信息储存到视频表，作者信息储存到作者表。新数据标记未同步历史数据，未关注
func getHistory(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			log.ErrorLog.Println(string(debug.Stack()))
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				log.ErrorLog.Println("执行报错，重新添加历史数据爬取")
				var err error
				_, err = wheel.AppendOnceFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: oneTicket})
				if err != nil {
					log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	var err error
	VideoHistoryChan := make(chan models.Video)
	VideoHistoryCloseChan := make(chan models.TaskClose)
	go bilibili.Spider.GetVideoHistoryList(VideoHistoryChan, VideoHistoryCloseChan)
	for {
		select {
		case videoInfo := <-VideoHistoryChan:
			videoInfo.UpdateVideo()
		case closeInfo := <-VideoHistoryCloseChan:
			if closeInfo.WebSite == bilibili.Spider.GetWebSiteName().WebName {
				for _, info := range closeInfo.Data {
					err = models.SaveSpiderParamByUserId(info.UserId, "history_baseline", info.EndBaseLine)
					if err != nil {
						log.ErrorLog.Printf("保存dynamic_baseline失败：%s\n", err.Error())
					}
				}
			}
			_, err = wheel.AppendOnceFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: historyRunTime()})
			if err != nil {
				log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				return
			}
			return
		}
	}

}

// 更新收藏夹列表，喝订阅的合集列表，新创建的同步视频数据
func updateCollectList(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
				if err != nil {
					log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	web := models.WebSite{
		WebName: "bilibili",
	}
	err := web.GetOrCreate()
	if err != nil {
		log.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
		return
	}
	newCollectList := bilibili.Spider.GetCollectList()
	var (
		videoCollectList []models.CollectVideo
		have             bool
	)
	for _, collectId := range newCollectList.Collect {
		models.GormDB.Table("collect_video").Where("collect_id=?", collectId.CollectId).Find(&videoCollectList)
		if int64(len(videoCollectList)) != collectId.CollectNumber {
			for _, info := range bilibili.Spider.GetCollectAllVideo(collectId.BvId, 0) {
				uploadTime := time.Unix(info.Ctime, 0)
				vi := models.Video{
					WebSiteId:  web.Id,
					Title:      info.Title,
					VideoDesc:  info.Intro,
					Duration:   info.Duration,
					Uuid:       info.Bvid,
					CoverUrl:   info.Cover,
					UploadTime: &uploadTime,
					Authors: []models.VideoAuthor{
						{Uuid: info.BvId, AuthorUUID: strconv.Itoa(info.Upper.Mid)},
					},
					StructAuthor: []models.Author{
						{
							WebSiteId:    web.Id,
							AuthorName:   info.Upper.Name,
							AuthorWebUid: strconv.Itoa(info.Upper.Mid),
							Avatar:       info.Upper.Face,
						},
					},
				}
				vi.UpdateVideo()
				mtine := time.Unix(info.FavTime, 0)
				for _, videoCollectInfo := range videoCollectList {
					if videoCollectInfo.VideoId == vi.Id {
						have = true
						break
					}
				}
				if !have {
					models.CollectVideo{
						CollectId: collectId.CollectId,
						VideoId:   vi.Id,
						Mtime:     &mtine,
					}.Save()
				}
			}
			time.Sleep(time.Second * 5)
		}
	}
	for _, collectId := range newCollectList.Season {
		models.GormDB.Table("collect_video").Where("collect_id=?", collectId.CollectId).Find(&videoCollectList)
		if int64(len(videoCollectList)) != collectId.CollectNumber {
			for _, info := range bilibili.Spider.GetSeasonAllVideo(collectId.BvId) {
				author := models.Author{
					WebSiteId:    web.Id,
					AuthorWebUid: strconv.Itoa(info.Upper.Mid),
					AuthorName:   info.Upper.Name,
					Follow:       false,
					Crawl:        true,
				}
				err = author.GetOrCreate()
				if err != nil {
					log.ErrorLog.Printf("获取作者信息失败：%s\n", err.Error())
					continue
				}
				vi := models.Video{
					WebSiteId: web.Id,
					Title:     info.Title,
					Duration:  info.Duration,
					Uuid:      info.Bvid,
					CoverUrl:  info.Cover,
					Authors: []models.VideoAuthor{
						{Uuid: info.Bvid, AuthorUUID: strconv.Itoa(info.Upper.Mid)},
					},
					StructAuthor: []models.Author{
						{
							WebSiteId:    web.Id,
							AuthorName:   info.Upper.Name,
							AuthorWebUid: strconv.Itoa(info.Upper.Mid),
						},
					},
				}
				vi.UpdateVideo()
				for _, videoCollectInfo := range videoCollectList {
					if videoCollectInfo.VideoId == vi.Id {
						have = true
						break
					}
				}
				if !have {
					models.CollectVideo{
						CollectId: collectId.CollectId,
						VideoId:   vi.Id,
					}.Save()
				}
			}
			time.Sleep(time.Second * 5)
		}
	}

	_, err = wheel.AppendOnceFunc(updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + rand.Int63n(100)})
	if err != nil {
		log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

// 更新收藏夹视频列表，更新合集视频列表
func updateCollectVideoList(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(updateCollectVideoList, nil, "updateCollectVideoSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(twelveTicket, sixTime, twentyTime)})
				if err != nil {
					log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	var (
		collectId int64
		videoList []string
	)

	queryResult := models.GetAllCollectVideo()
	collectVideoGroup := make(map[int64][]string)
	for _, info := range queryResult {
		collectVideoGroup[info.CollectId] = append(collectVideoGroup[info.CollectId], info.Uuid)
	}

	waitUpdateList := make([]int64, 0)
	for collectId, videoList = range collectVideoGroup {
		r := bilibili.GetCollectVideoList(collectId, "干煸花椰菜")
		// countList 和r.Data中的BvId对比，找不同的值
		for _, info := range r.Data {
			if !utils.InArray(info.BvId, videoList) {
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
	web.GetOrCreate()
	for _, collectId = range waitUpdateList {
		r := bilibili.Spider.GetCollectAllVideo(collectId, 1)
		videoList = collectVideoGroup[collectId]
		for _, info := range r {
			if !utils.InArray(info.BvId, videoList) {
				vi := models.Video{
					WebSiteId: web.Id,
					Title:     info.Title,
					Duration:  info.Duration,
					Uuid:      info.Bvid,
					CoverUrl:  info.Cover,
					Authors: []models.VideoAuthor{
						{Uuid: info.BvId, AuthorUUID: strconv.Itoa(info.Upper.Mid)},
					},
					StructAuthor: []models.Author{
						{
							WebSiteId:    web.Id,
							AuthorWebUid: strconv.Itoa(info.Upper.Mid),
							AuthorName:   info.Upper.Name,
						},
					},
				}
				vi.UpdateVideo()
				models.CollectVideo{
					CollectId: collectId,
					VideoId:   vi.Id,
				}.Save()
			}
		}
	}

	_, err := wheel.AppendOnceFunc(updateCollectVideoList, nil, "updateCollectVideoSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(twelveTicket, sixTime, twentyTime)})
	if err != nil {
		log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

// 获取稍后观看列表信息
func getWatchLaterList(interface{}) {

}

// 同步关注信息
func updateFollowInfo(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
				if err != nil {
					log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	web := models.WebSite{
		WebName: "bilibili",
	}
	err := web.GetOrCreate()
	if err != nil {
		log.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
		return
	}
	resultChan := make(chan baseStruct.FollowInfo)
	closeChan := make(chan int64)
	go bilibili.Spider.GetFollowingList(resultChan, closeChan)

	var (
		followList           map[int64]map[string]int64
		followInfo           baseStruct.FollowInfo
		closeSign            bool
		insertFollowRelation bool
		receiveUserId        map[int64]bool
	)
	followList = models.GetFollowList(web.Id)
	receiveUserId = make(map[int64]bool)

	for {
		select {
		case followInfo = <-resultChan:
			receiveUserId[followInfo.UserId] = true
			insertFollowRelation = false
			userFollowMap, ok := followList[followInfo.UserId]
			if !ok {
				insertFollowRelation = true
			}
			_, ok = userFollowMap[followInfo.AuthorUUID]
			if !ok {
				insertFollowRelation = true
			}
			if insertFollowRelation {
				author := models.Author{
					WebSiteId:    web.Id,
					AuthorWebUid: followInfo.AuthorUUID,
					AuthorName:   followInfo.AuthorName,
					Avatar:       followInfo.Avatar,
					AuthorDesc:   followInfo.AuthorDesc,
					Follow:       true,
					FollowTime:   followInfo.FollowTime,
				}
				err = author.UpdateOrCreate()
				if err != nil {
					log.ErrorLog.Printf("更新作者信息失败：%s\n", err.Error())
					continue
				}
				f := models.Follow{
					WebSiteId:  followInfo.WebSiteId,
					AuthorId:   author.Id,
					UserId:     followInfo.UserId,
					FollowTime: followInfo.FollowTime,
				}
				models.GormDB.Create(&f)
			}
			// 从userFollowMap列表中删除这个作者
			delete(userFollowMap, followInfo.AuthorUUID)
		case <-closeChan:
			closeSign = true
		}
		if closeSign {
			break
		}
	}
	// followList中剩下的作者，标记为未关注
	if closeSign {
		var deleteFollowId []int64
		for userId, authorMap := range followList {
			_, ok := receiveUserId[userId]
			if !ok {
				continue
			}
			for _, followId := range authorMap {
				deleteFollowId = append(deleteFollowId, followId)
			}
		}
		if len(deleteFollowId) > 0 {
			models.GormDB.Delete(&models.Follow{}, deleteFollowId)
		}
	}

	_, err = wheel.AppendOnceFunc(updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + rand.Int63n(100)})
	if err != nil {
		log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

func historyRunTime() int64 {
	// 周一到周五 5:00-9:30、11:30-13:40 、16:30-12:00 这些时间段每一小时运行一次，其他时间每2小时运行一次
	// 周末 1:00-7:00 不运行，其他时间每小时运行一次
	nowTime := time.Now()
	zeroTime := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location())
	timeGap := int64(nowTime.Sub(zeroTime) / 1000000000)
	if nowTime.Weekday() == time.Saturday || nowTime.Weekday() == time.Sunday {
		if 3600 < timeGap && timeGap < 25200 {
			return 25200 - timeGap + rand.Int63n(100)
		}
		return 3600 + rand.Int63n(100)
	}
	if 18000 < timeGap && timeGap < 34200 {
		return 3600 + rand.Int63n(100)
	}
	if 41400 < timeGap && timeGap < 46800 {
		return 3600 + rand.Int63n(100)
	}
	if 59400 < timeGap && timeGap < 86400 {
		return 3600 + rand.Int63n(100)
	}
	return 7200 + rand.Int63n(100)
}

func loadProxyInfo(interface{}) {
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		println(err.Error())
		return
	}
	configFile := &utils.Config{}
	err = json.Unmarshal(fileData, configFile)
	if err != nil {
		println(err.Error())
		return
	}
	config.Proxy = configFile.Proxy
}

// 对没有关注的作者爬取最新的视频信息
func updateAuthorVideoList(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				_, err := wheel.AppendOnceFunc(updateAuthorVideoList, nil, "updateAuthorVideoSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(twelveTicket, sixTime, twentyTime)})
				if err != nil {
					log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	web := models.WebSite{
		WebName: "bilibili",
	}
	err := web.GetOrCreate()
	if err != nil {
		log.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
		return
	}
	authorList := models.GetCrawlAuthorList(web.Id)
	for _, author := range authorList {
		if author.Follow {
			continue
		}
		authorVideoPage := bilibili.Spider.GetAuthorVideoList(author.AuthorWebUid, 1, 2)
		for _, videoListInfo := range authorVideoPage {
			for _, videoInfo := range videoListInfo.Data.List.Vlist {
				uploadTime := time.Unix(videoInfo.Created, 0)
				authors := []models.VideoAuthor{}
				structAuthor := []models.Author{}
				// 是否是联合投稿 0：否 1：是
				if videoInfo.IsUnionVideo == 0 {
					authors = append(authors, models.VideoAuthor{Uuid: videoInfo.Bvid, AuthorUUID: author.AuthorWebUid})
					structAuthor = append(structAuthor, models.Author{
						WebSiteId:  web.Id,
						AuthorName: videoInfo.Author,
					})
				}
				vi := models.Video{
					WebSiteId:    web.Id,
					Title:        videoInfo.Title,
					VideoDesc:    videoInfo.Description,
					Duration:     bilibili.HourAndMinutesAndSecondsToSeconds(videoInfo.Length),
					Uuid:         videoInfo.Bvid,
					CoverUrl:     videoInfo.Pic,
					UploadTime:   &uploadTime,
					Authors:      authors,
					StructAuthor: structAuthor,
				}
				vi.UpdateVideo()
			}
		}
	}
}
