package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"math/rand"
	"os"
	"runtime/debug"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
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
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
	historyTaskId   int64
	dataPath        string
	config          *utils.Config
)

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList(chan<- models.Video, chan<- baseStruct.TaskClose, int64)
}

type Spider struct {
	interval int64
}

func readConfig() error {
	fileData, err := os.ReadFile("C:\\Code\\GO\\videoDynamicSpider\\cmd\\spider\\config.json")
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
	err = readConfig()
	if err != nil {
		os.Exit(2)
		return
	}
	if config.DataPath != "" {
		dataPath = config.DataPath
		baseStruct.RootPath = config.DataPath
	} else {
		dataPath = baseStruct.RootPath
	}
	utils.InitLog(dataPath)

	models.InitDB(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User, config.DB.Password, config.DB.HOST, config.DB.Port, config.DB.DatabaseName), true)
	wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
		IsRun: false,
		Log:   utils.TimeWheelLog,
	})
	rand.Seed(time.Now().UnixNano())
	utils.Info.Println("初始化完成：", time.Now().Format("2006.01.02 15:04:05"))
}

func main() {
	videoCollection = []VideoCollection{
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
	historyTaskId, err = wheel.AppendOnceFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: 10})
	if err != nil {
		return
	}
	//_, err = wheel.AppendOnceFunc(updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + 120})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendOnceFunc(updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
	//if err != nil {
	//	return
	//}
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

	if leftBorder > timeGap || timeGap > rightBorder {
		// 早上六点之前晚上八点之后，不再执行。六点之后才执行
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
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()

	videoResultChan := make(chan models.Video)
	closeChan := make(chan baseStruct.TaskClose)
	runWebSite := make([]string, 0)

	for _, v := range videoCollection {
		webSite := models.WebSite{}
		models.GormDB.Where("web_name=?", v.GetWebSiteName().WebName).First(&webSite)
		if webSite.Id == 0 {
			utils.ErrorLog.Printf("%s站点在数据库中找不到对应的数据。", v.GetWebSiteName().WebName)
			continue
		}
		go v.GetVideoList(videoResultChan, closeChan, webSite.Id)
		runWebSite = append(runWebSite, v.GetWebSiteName().WebName)
	}

	var (
		videoInfo models.Video
		closeInfo baseStruct.TaskClose
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
			if closeInfo.WebSite == "bilibili" && closeInfo.Code > 0 {

			}
		}
		if len(runWebSite) == 0 {
			break
		}
	}
	close(closeChan)
	close(videoResultChan)

	_, err = wheel.AppendOnceFunc(s.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(defaultTicket, sixTime, twentyTime)})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

// 定时抓取历史记录 历史数据，同步到以观看表,视频信息储存到视频表，作者信息储存到作者表。新数据标记未同步历史数据，未关注
func getHistory(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			utils.ErrorLog.Println(string(debug.Stack()))
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				utils.ErrorLog.Println("执行报错，重新添加历史数据爬取")
				var err error
				historyTaskId, err = wheel.AppendOnceFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: oneTicket})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()
	utils.Info.Printf("历史任务执行id：%d\n", historyTaskId)
	bilibili.SaveVideoHistoryList()
	var err error
	historyTaskId, err = wheel.AppendOnceFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: historyRunTime()})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
	/*
		baseLine := models.GetHistoryBaseLine()
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
		VideoHistoryChan := make(chan models.Video)
		VideoHistoryCloseChan := make(chan int64)
		go bilibili.Spider.GetVideoHistoryList(lastHistoryTimestamp, VideoHistoryChan, VideoHistoryCloseChan)
		website := models.WebSite{WebName: "bilibili"}
		err = website.GetOrCreate()
		if err != nil {
			utils.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
			return
		}
		for {
			select {
			case videoInfo := <-VideoHistoryChan:
				videoInfo.UpdateVideo()
			case newestTimestamp := <-VideoHistoryCloseChan:
				models.SaveHistoryBaseLine(strconv.FormatInt(newestTimestamp, 10))
				historyTaskId, err = wheel.AppendOnceFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: historyRunTime()})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
					return
				}
				return
			}
		}

	*/
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
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
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
		utils.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
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
					utils.ErrorLog.Printf("获取作者信息失败：%s\n", err.Error())
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
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
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
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
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
		r := bilibili.GetCollectVideoList(collectId)
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
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
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
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
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
		utils.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
		return
	}
	resultChan := make(chan bilibili.FollowingUP)
	closeChan := make(chan int64)
	go bilibili.Spider.GetFollowingList(resultChan, closeChan)

	var (
		followList []models.Author
		upInfo     bilibili.FollowingUP
		closeSign  bool
	)
	followList = models.GetFollowList(web.Id)

	for {
		select {
		case upInfo = <-resultChan:
			followTime := time.Unix(upInfo.Mtime, 0)
			author := models.Author{
				WebSiteId:    web.Id,
				AuthorWebUid: strconv.FormatInt(upInfo.Mid, 10),
				AuthorName:   upInfo.Uname,
				Avatar:       upInfo.Face,
				AuthorDesc:   upInfo.Sign,
				Follow:       true,
				FollowTime:   &followTime,
			}
			err = author.UpdateOrCreate()
			if err != nil {
				utils.ErrorLog.Printf("更新作者信息失败：%s\n", err.Error())
				continue
			}
			// 从followList列表中删除这个作者
			for index, v := range followList {
				if v.AuthorWebUid == author.AuthorWebUid {
					followList = append(followList[:index], followList[index+1:]...)
					break
				}
			}
		case <-closeChan:
			closeSign = true
		}
		if closeSign {
			break
		}
	}
	// followList中剩下的作者，标记为未关注
	if closeSign {
		for _, v := range followList {
			v.Follow = false
			err = v.UpdateOrCreate()
			if err != nil {
				utils.ErrorLog.Printf("取消关注作者信息失败：%s\n", err.Error())
				continue
			}
		}
	}
	_, err = wheel.AppendOnceFunc(updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + rand.Int63n(100)})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
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
//func updateAuthorVideoList(interface{}) {
//	defer func() {
//		panicErr := recover()
//		if panicErr != nil {
//			_, ok := panicErr.(utils.DBFileLock)
//			if ok {
//				_, err := wheel.AppendOnceFunc(updateAuthorVideoList, nil, "updateAuthorVideoSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(twelveTicket, sixTime, twentyTime)})
//				if err != nil {
//					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
//				}
//				return
//			}
//			panic(panicErr)
//		}
//	}()
//
//	web := models.WebSite{
//		WebName: "bilibili",
//	}
//	err := web.GetOrCreate()
//	if err != nil {
//		utils.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
//		return
//	}
//	authorList := models.GetCrawlAuthorList(web.Id)
//	for _, author := range authorList {
//		if author.Follow {
//			continue
//		}
//		authorInfo := bilibili.Spider.GetAuthorInfo(author.AuthorWebUid)
//		if authorInfo.Mid == 0 {
//			continue
//		}
//		author.AuthorName = authorInfo.Name
//		author.Avatar = authorInfo.Face
//		author.AuthorDesc = authorInfo.Sign
//		author.Follow = true
//		author.FollowTime = time.Now()
//		err = author.UpdateOrCreate()
//		if err != nil {
//			utils.ErrorLog.Printf("更新作者信息失败：%s\n", err.Error())
//			continue
//		}
//		for _, videoInfo := range bilibili.Spider.GetAuthorVideoList(author.AuthorWebUid) {
//			uploadTime := time.Unix(videoInfo.Pubdate, 0)
//			vi := models.Video{
//				WebSiteId:  web.Id,
//				Title:      videoInfo.Title,
//				VideoDesc:  videoInfo.Desc,
//				Duration:   videoInfo.Duration,
//				Uuid:       videoInfo.Bvid,
//				CoverUrl:   videoInfo.Pic,
//				UploadTime: &uploadTime,
//				Authors: []models.VideoAuthor{
//					{Uuid: videoInfo.Bvid, AuthorUUID: author.AuthorWebUid},
//				},
//				StructAuthor: []models.Author{
//					{
//						WebSiteId:  web.Id,
//						AuthorName: authorInfo.Name,
//					},
//				},
//			}
//		}
//	}
//}
