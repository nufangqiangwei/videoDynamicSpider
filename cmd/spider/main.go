package main

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"io"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/grpcDiscovery/DHT"
	"videoDynamicAcquisition/grpcDiscovery/redis"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
	webSiteGRPC "videoDynamicAcquisition/proto"
	"videoDynamicAcquisition/proxy"
	"videoDynamicAcquisition/utils"
)

// 几尺戏台上，演尽痴心梦。
const (
	oneMinute     = 60
	defaultTicket = 60 * 5
	oneTicket     = 60 * 60
	sixTime       = oneTicket * 6
	twentyTime    = oneTicket * 20
	twelveTicket  = oneTicket * 12
	configPath    = "./spiderConfig.json"
	dialTimeout   = time.Minute * 10
)

var (
	spiderWebSit            []models.VideoCollection
	wheel                   *timeWheel.TimeWheel
	config                  *utils.Config
	wheelLog                log.LogInputFile
	databaseLog             log.LogInputFile
	waitUpdateVideoInfoChan chan models.Video
	webSiteManage           webSite
	grpcServer              map[string]webSiteGRPC.WebSiteServiceClient
)

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

	// 初始化网站数据
	webSiteManage = webSite{}
	webSiteManage.init()

	// 初始化代理
	proxy.Init()

	// 初始化grpc客户端
	redisDiscovery.OpenRedis()
	err = getWebSiteGrpcClient()
	if err != nil {
		panic(err)
	}

	// 初始化通道
	waitUpdateVideoInfoChan = make(chan models.Video, 100)
	// 初始化定时器
	wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
		IsRun: false,
		Log:   wheelLog.WriterObject,
	})
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//rand.Seed(time.Now().UnixNano())
	log.Info.Println("初始化完成：", time.Now().Format("2006.01.02 15:04:05"))
}

func initWebSiteSpider() {
	//websiteMap := models.GetAllWebSite()
	//var (
	//	dynamicBaseLine map[string]int64
	//	historyBaseLine map[string]int64
	//)
	//dynamicBaseLine = make(map[string]int64)
	//historyBaseLine = make(map[string]int64)
	cookies.RangeCookiesMap(func(webSiteName, userName string, userCookie *cookies.UserCookie) {
		if strings.HasPrefix(userName, cookies.Tourists) {
			return
		}
		userInfo, err := models.GetAuthorByUserName(webSiteName, userName)
		if err != nil {
			log.ErrorLog.Printf("获取%s用户id失败：%s\n", userName, err.Error())
			return
		}
		userCookie.SetDBPrimaryKeyId(userInfo.Id)
		userCookie.SetWebSiteId(userInfo.WebSiteId)

		//if webSiteName == bilibili.Spider.GetWebSiteName().WebName {
		//	var (
		//		result            []models.UserSpiderParams
		//		err               error
		//		intLatestBaseline int64
		//	)
		//	models.GormDB.Model(&models.UserSpiderParams{}).Where("author_id = ?", userId).Find(&result)
		//	for _, rowData := range result {
		//		intLatestBaseline, err = strconv.ParseInt(rowData.Values, 10, 64)
		//		if err != nil {
		//			log.ErrorLog.Printf("转换%s的history_baseline失败：%s\n", userName, err.Error())
		//			continue
		//		}
		//		if rowData.KeyName == "dynamic_baseline" {
		//			dynamicBaseLine[userName] = intLatestBaseline
		//		}
		//		if rowData.KeyName == "history_baseline" {
		//			historyBaseLine[userName] = intLatestBaseline
		//		}
		//	}
		//}
	})
	//bilibili.Spider.Init(dynamicBaseLine, historyBaseLine, websiteMap[webSiteName].Id)
}

func main() {
	var err error
	_, err = wheel.AppendCycleFunc(func(interface{}) {
		e := getWebSiteGrpcClient()
		if e != nil {
			log.ErrorLog.Printf("getWebSiteGrpcClient error: %s", e.Error())
		}
	}, nil, "flushGrpcClient", timeWheel.Crontab{ExpiredTime: defaultTicket})
	if err != nil {
		return
	}
	_, err = wheel.AppendOnceFunc(getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: defaultTicket})
	if err != nil {
		return
	}
	_, err = wheel.AppendCycleFunc(getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: oneTicket})
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
	go func() {
		time.Sleep(time.Minute)
		getHistory(nil)
	}()
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

func getDynamic(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			err, ok := panicErr.(runtime.Error)
			if ok && err.Error() == "invalid memory address or nil pointer dereference" {
				log.ErrorLog.Printf("出现空指针错误：定时器时间是%s", wheel.PrintTime())
			}
			panic(panicErr)
		}
	}()
	defer func() {
		nextRunTime := arrangeRunTime(defaultTicket, sixTime, twentyTime)
		log.Info.Printf("%d秒后再次抓取动态", nextRunTime)
		_, err := wheel.AppendOnceFunc(getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: nextRunTime})
		if err != nil {
			log.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
			return
		}
	}()
	var (
		dynamicResult chan *webSiteGRPC.VideoInfoResponse
		runWebSite    = make([]string, 0)
	)
	dynamicResult = make(chan *webSiteGRPC.VideoInfoResponse, 10)
	for websiteServerName, server := range grpcServer {
		for userName, userCookie := range cookies.GetWebSiteUser(websiteServerName) {
			go func(websiteServerName, userName string, userCookie cookies.UserCookie) {
				cookieMap := userCookie.GetCookiesDictToLowerKey()
				cookieMap["requestUserName"] = userName
				res, err := server.GetUserFollowUpdate(context.Background(), &webSiteGRPC.UserInfo{
					Cookies:        cookieMap,
					LastUpdateTime: models.GetDynamicBaseline(userCookie.GetDBPrimaryKeyId()),
				})
				if err != nil {
					log.ErrorLog.Printf("获取%s的用户关注更新失败：%s", websiteServerName, err.Error())
					dynamicResult <- &webSiteGRPC.VideoInfoResponse{
						ErrorCode:       500,
						ErrorMsg:        err.Error(),
						WebSiteName:     websiteServerName,
						RequestUserName: userName,
						WebSiteId:       userCookie.GetWebSiteId(),
						RequestUserId:   userCookie.GetDBPrimaryKeyId(),
					}
					return
				}
				for {
					vvi, err := res.Recv()
					if err == io.EOF {
						return
					}
					if err != nil {
						vvi = &webSiteGRPC.VideoInfoResponse{
							ErrorCode:       500,
							ErrorMsg:        err.Error(),
							WebSiteName:     websiteServerName,
							RequestUserName: userName,
							WebSiteId:       userCookie.GetWebSiteId(),
							RequestUserId:   userCookie.GetDBPrimaryKeyId(),
						}
						return
					}
					vvi.WebSiteName = websiteServerName
					vvi.RequestUserName = userName
					vvi.WebSiteId = userCookie.GetWebSiteId()
					vvi.RequestUserId = userCookie.GetDBPrimaryKeyId()
					dynamicResult <- vvi
				}
			}(websiteServerName, userName, *userCookie)
			runWebSite = append(runWebSite, fmt.Sprintf("%s-%s", websiteServerName, userName))
		}
	}
	var pushTime time.Time
	videoNumber := 0
	for {
		select {
		case videoInfoResponse := <-dynamicResult:
			if videoInfoResponse.ErrorCode != 0 {
				if videoInfoResponse.ErrorCode == 200 {
					log.ErrorLog.Printf("获取动态完成：%s 用户获取到%s\n", videoInfoResponse.RequestUserName, videoInfoResponse.ErrorMsg)
				} else {
					log.ErrorLog.Printf("获取动态失败：%s\n", videoInfoResponse.ErrorMsg)
				}

				userInfo := fmt.Sprintf("%s-%s", videoInfoResponse.WebSiteName, videoInfoResponse.RequestUserName)
				for index, v := range runWebSite {
					if v == userInfo {
						runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
						break
					}
				}
				if len(runWebSite) == 0 {
					log.Info.Printf("所有网站已经获取动态完毕，共获取%d个动态，退出\n", videoNumber)
					return
				}
			} else if videoInfoResponse.ErrorCode == 0 {
				pushTime = time.Unix(videoInfoResponse.UpdateTime, 0)
				video := models.Video{
					WebSiteId:  videoInfoResponse.WebSiteId,
					Title:      videoInfoResponse.Title,
					VideoDesc:  videoInfoResponse.Desc,
					Duration:   int(videoInfoResponse.Duration),
					Uuid:       videoInfoResponse.Uid,
					CoverUrl:   videoInfoResponse.Cover,
					UploadTime: &pushTime,
					Authors: []models.VideoAuthor{
						{Contribute: "UP主", AuthorUUID: videoInfoResponse.Authors[0].Uid},
					},
					StructAuthor: []models.Author{
						{
							WebSiteId:    videoInfoResponse.WebSiteId,
							AuthorName:   videoInfoResponse.Authors[0].Name,
							AuthorWebUid: videoInfoResponse.Authors[0].Uid,
							Avatar:       videoInfoResponse.Authors[0].Author,
						},
					},
				}
				video.UpdateVideo()
				videoNumber++
			}
		}
	}
}

// 定时抓取历史记录 历史数据，同步到以观看表,视频信息储存到视频表，作者信息储存到作者表。新数据标记未同步历史数据，未关注
func getHistory(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			err, ok := panicErr.(runtime.Error)
			if ok && err.Error() == "invalid memory address or nil pointer dereference" {
				log.ErrorLog.Printf("出现空指针错误：定时器时间是%s", wheel.PrintTime())
			}
			panic(panicErr)
		}
	}()

	var (
		historyResult chan *webSiteGRPC.VideoInfoResponse
		runWebSite    = make([]string, 0)
	)
	historyResult = make(chan *webSiteGRPC.VideoInfoResponse, 50)
	for websiteServerName, server := range grpcServer {
		for userName, userCookie := range cookies.GetWebSiteUser(websiteServerName) {
			go func(websiteServerName, userName string, userCookie cookies.UserCookie) {
				cookieMap := userCookie.GetCookiesDictToLowerKey()
				cookieMap["requestUserName"] = userName
				res, err := server.GetUserViewHistory(context.Background(), &webSiteGRPC.UserInfo{
					Cookies:         cookieMap,
					LastHistoryTime: models.GetHistoryBaseline(userCookie.GetDBPrimaryKeyId()),
				})
				if err != nil {
					log.ErrorLog.Printf("GRPC服务端%s获取历史记录失败:%s\n", userName, err.Error())
					historyResult <- &webSiteGRPC.VideoInfoResponse{
						ErrorCode:       500,
						ErrorMsg:        err.Error(),
						WebSiteName:     websiteServerName,
						RequestUserName: userName,
						WebSiteId:       userCookie.GetWebSiteId(),
						RequestUserId:   userCookie.GetDBPrimaryKeyId(),
					}
					return
				}
				for {
					vvi, err := res.Recv()
					if err == io.EOF {
						return
					}
					if err != nil {
						log.ErrorLog.Printf("GRPC服务端%s获取历史记录流通道错误:%s\n", userName, err.Error())
						vvi = &webSiteGRPC.VideoInfoResponse{
							ErrorCode:       500,
							ErrorMsg:        err.Error(),
							WebSiteName:     websiteServerName,
							RequestUserName: userName,
							WebSiteId:       userCookie.GetWebSiteId(),
							RequestUserId:   userCookie.GetDBPrimaryKeyId(),
						}
					}
					vvi.WebSiteName = websiteServerName
					vvi.RequestUserName = userName
					vvi.WebSiteId = userCookie.GetWebSiteId()
					vvi.RequestUserId = userCookie.GetDBPrimaryKeyId()

					historyResult <- vvi
					if vvi.ErrorCode != 0 {
						log.ErrorLog.Printf("%s获取历史状态错误:%s\n", userName, vvi.ErrorMsg)
						break
					}
				}
			}(websiteServerName, userName, *userCookie)
			runWebSite = append(runWebSite, fmt.Sprintf("%s-%s", websiteServerName, userName))
		}
	}
	var viewTime time.Time
	videoNumber := 0
	for vi := range historyResult {
		if vi.ErrorCode == 0 {
			viewTime = time.Unix(vi.ViewInfo.ViewTime, 0)
			video := models.Video{
				WebSiteId: vi.WebSiteId,
				Title:     vi.Title,
				VideoDesc: vi.Desc,
				Duration:  int(vi.Duration),
				Uuid:      vi.Uid,
				CoverUrl:  vi.Cover,
				Authors: []models.VideoAuthor{
					{
						Contribute: "UP主",
						AuthorUUID: vi.Authors[0].Uid,
						Uuid:       vi.Uid,
					},
				},
				StructAuthor: []models.Author{
					{
						AuthorName:   vi.Authors[0].Name,
						AuthorWebUid: vi.Authors[0].Uid,
						Avatar:       vi.Authors[0].Avatar,
					},
				},
				ViewHistory: []models.VideoHistory{
					{
						ViewTime: viewTime,
						Duration: int(vi.ViewInfo.ViewDuration),
						WebUUID:  vi.Uid,
						AuthorId: vi.RequestUserId,
					},
				},
			}
			video.UpdateVideo()
			videoNumber++
		} else {
			userInfo := fmt.Sprintf("%s-%s", vi.WebSiteName, vi.RequestUserName)
			for index, v := range runWebSite {
				if v == userInfo {
					runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
					break
				}
			}
			if len(runWebSite) == 0 {
				log.Info.Printf("全部获取完成，共获取%d个视频\n", videoNumber)
				return
			}
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
	var err error
	web := webSiteManage.getWebSiteByName("bilibili")
	newCollectList := bilibili.Spider.GetCollectList()
	var (
		videoCollectList []models.CollectVideo
		have             bool
	)
	for _, collectId := range newCollectList.Collect {
		models.GormDB.Table("collect_video").Where("collect_id=?", collectId.Id).Find(&videoCollectList)
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
					CollectId: collectId.Id,
					VideoId:   vi.Id,
					Mtime:     &mtine,
				}.Save()
			}
		}
		time.Sleep(time.Second * 5)
	}
	for _, collectId := range newCollectList.Season {
		models.GormDB.Table("collect_video").Where("collect_id=?", collectId.Id).Find(&videoCollectList)
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
					CollectId: collectId.Id,
					VideoId:   vi.Id,
				}.Save()
			}
		}
		time.Sleep(time.Second * 5)
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
	var err error
	web := webSiteManage.getWebSiteByName("bilibili")
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
		fileObject := utils.WriteFile{
			FileName: func(s string) string {
				return "取消关注信息.txt"
			},
			FolderPrefix: []string{baseStruct.RootPath},
		}
		for userId, authorMap := range followList {
			_, ok := receiveUserId[userId]
			if !ok {
				continue
			}
			if len(authorMap) > 0 {
				fileObject.WriteLine([]byte(strconv.FormatInt(userId, 10)))
			}
			for _, followId := range authorMap {
				println(followId)
				deleteFollowId = append(deleteFollowId, followId)
				fileObject.Write([]byte{32, 32, 32, 32})
				fileObject.WriteLine([]byte(strconv.FormatInt(followId, 10)))
			}
		}
		fileObject.Close()
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
func updateWebSiteUserFollowInfo(site webSite, cookie cookies.UserCookie) {

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

// 待更新的视频信息
func waitUpdateVideo(interface{}) {
	for v := range waitUpdateVideoInfoChan {
		go func(videoInfo models.Video) {
			data, _ := bilibili.GetVideoDetailByByte(videoInfo.Uuid)
			response := bilibili.VideoDetailResponse{}
			err := response.BindJSON(data)
			if err != nil {
				log.ErrorLog.Println(err)
				return
			}
			var uploadTime time.Time
			uploadTime = time.Unix(response.Data.View.Ctime, 0)
			DatabaseAuthorInfo := []models.Author{
				{
					WebSiteId:    videoInfo.WebSiteId,
					AuthorName:   response.Data.View.Owner.Name,
					AuthorWebUid: strconv.FormatInt(response.Data.View.Owner.Mid, 10),
					Avatar:       response.Data.View.Owner.Face,
				},
			}
			if len(response.Data.View.Staff) > 0 {
				for _, staff := range response.Data.View.Staff {
					DatabaseAuthorInfo = append(DatabaseAuthorInfo, models.Author{
						WebSiteId:    videoInfo.WebSiteId,
						AuthorName:   staff.Name,
						AuthorWebUid: strconv.FormatInt(staff.Mid, 10),
						Avatar:       staff.Face,
						FollowNumber: staff.Follower,
					})
				}
			}

			result := models.Video{
				WebSiteId:    videoInfo.WebSiteId,
				Title:        response.Data.View.Title,
				VideoDesc:    response.Data.View.Desc,
				Duration:     response.Data.View.Duration,
				Uuid:         response.Data.View.Bvid,
				CoverUrl:     response.Data.View.Pic,
				UploadTime:   &uploadTime,
				StructAuthor: DatabaseAuthorInfo,
			}

			result.UpdateVideo()
		}(v)
	}
}

func getAuthorFirstVideo(interface{}) {
	authorList := []models.Author{}
	/* select a.* from author a inner join follow f on f.author_id = a.id inner join user_web_site_account ua on ua.author_id = f.user_id where ua.user_id = 11 order by f.follow_time desc*/
	err := models.GormDB.Model(&models.Author{}).Joins("inner join follow f on f.author_id = author.id ").Joins(
		"inner join user_web_site_account ua on ua.author_id = f.user_id").Where("ua.user_id = 11").Order("f.follow_time desc").Limit(500).Offset(138).Find(&authorList).Error

	if err != nil {
		println(err.Error())
	}
	if len(authorList) == 0 {
		return
	}
	userList := cookies.GetWeb("bilibili")
	var (
		video    models.Video
		pushTime time.Time
	)
	for index, authorObject := range authorList {
		println(index, " 作者：", authorObject.AuthorName)
		response := bilibili.GetAuthorDynamic(authorObject.AuthorWebUid, "", userList.PickUser())
		if response == nil {
			return
		}
		for _, dynmaicInfo := range response.Data.Items {
			switch dynmaicInfo.Type {
			case bilibili.DynamicInfoType.DynamicTypeAv:
				fmt.Printf("视频标题：%s\n", dynmaicInfo.Modules.ModuleDynamic.Major.Archive.Title)
				pushTime = time.Unix(dynmaicInfo.Modules.ModuleAuthor.PubTs, 0)
				video = models.Video{
					WebSiteId:  1,
					Title:      dynmaicInfo.Modules.ModuleDynamic.Major.Archive.Title,
					VideoDesc:  dynmaicInfo.Modules.ModuleDynamic.Major.Archive.Desc,
					Duration:   bilibili.HourAndMinutesAndSecondsToSeconds(dynmaicInfo.Modules.ModuleDynamic.Major.Archive.DurationText),
					Uuid:       dynmaicInfo.Modules.ModuleDynamic.Major.Archive.Bvid,
					Url:        dynmaicInfo.Modules.ModuleDynamic.Major.Archive.JumpUrl,
					CoverUrl:   dynmaicInfo.Modules.ModuleDynamic.Major.Archive.Cover,
					UploadTime: &pushTime,
					Authors: []models.VideoAuthor{
						{Contribute: "UP主", AuthorUUID: strconv.Itoa(dynmaicInfo.Modules.ModuleAuthor.Mid)},
					},
					StructAuthor: []models.Author{
						{
							WebSiteId:    1,
							AuthorName:   dynmaicInfo.Modules.ModuleAuthor.Name,
							AuthorWebUid: strconv.Itoa(dynmaicInfo.Modules.ModuleAuthor.Mid),
							Avatar:       dynmaicInfo.Modules.ModuleAuthor.Face,
						},
					},
				}
				video.UpdateVideo()
			}
		}
		time.Sleep(time.Second)
	}
}

// 获取热门视频
func getHotVideo(interface{}) {
	var resultChan = make(chan models.Video, 10)
	var closeChan = make(chan int64)
	var closeSign bool
	var closeTaskNumber int
	for _, spider := range spiderWebSit {
		go spider.GetHotVideoList(resultChan, closeChan)
	}

	for {
		select {
		case vi := <-resultChan:
			vi.UpdateVideo()
		case <-closeChan:
			closeTaskNumber++
			closeSign = len(spiderWebSit) == closeTaskNumber
		}
		if closeSign {
			break
		}
	}

}

func textUpdateVideoInfo() {
	videoList := []models.Video{}
	models.GormDB.Model(&models.Video{}).Where("upload_time > now() - interval 30 day").Scan(&videoList)
	println("待更新", len(videoList), "个视频")

	for _, videoInfo := range videoList {
		data, _ := bilibili.GetVideoDetailByByte(videoInfo.Uuid)
		response := bilibili.VideoDetailResponse{}
		err := response.BindJSON(data)
		if err != nil {
			log.ErrorLog.Println(err)
			return
		}
		var uploadTime time.Time
		uploadTime = time.Unix(response.Data.View.Ctime, 0)
		DatabaseAuthorInfo := []models.Author{
			{
				WebSiteId:    videoInfo.WebSiteId,
				AuthorName:   response.Data.View.Owner.Name,
				AuthorWebUid: strconv.FormatInt(response.Data.View.Owner.Mid, 10),
				Avatar:       response.Data.View.Owner.Face,
			},
		}
		if len(response.Data.View.Staff) > 0 {
			for _, staff := range response.Data.View.Staff {
				DatabaseAuthorInfo = append(DatabaseAuthorInfo, models.Author{
					WebSiteId:    videoInfo.WebSiteId,
					AuthorName:   staff.Name,
					AuthorWebUid: strconv.FormatInt(staff.Mid, 10),
					Avatar:       staff.Face,
					FollowNumber: staff.Follower,
				})
			}
		}

		result := models.Video{
			WebSiteId:    videoInfo.WebSiteId,
			Title:        response.Data.View.Title,
			VideoDesc:    response.Data.View.Desc,
			Duration:     response.Data.View.Duration,
			Uuid:         response.Data.View.Bvid,
			CoverUrl:     response.Data.View.Pic,
			UploadTime:   &uploadTime,
			StructAuthor: DatabaseAuthorInfo,
			StructVideoPlayData: &models.VideoPlayData{
				View:       0,
				Danmaku:    0,
				Reply:      0,
				Favorite:   0,
				Coin:       0,
				Share:      0,
				NowRank:    0,
				HisRank:    0,
				Like:       0,
				Dislike:    0,
				Evaluation: "",
				CreateTime: time.Time{},
			},
		}

		result.UpdateVideo()
	}
}

func initGrpc() {
	dht, err := DHT.NewServiceDiscovery(&DHT.ServerConfig{
		ServerType:       "spider",
		SeedAddr:         "",
		AwaitRegister:    false,
		NodePort:         3190,
		ServerIp:         "127.0.0.1",
		GrpcSeverAddress: "127.0.0.1:3190",
	})
	if err != nil {
		log.ErrorLog.Println(err)
	}
	resolver.Register(dht)
}

func getWebSiteGrpcClient() error {
	serverList := redisDiscovery.GetWebSiteServer()
	if grpcServer == nil {
		grpcServer = make(map[string]webSiteGRPC.WebSiteServiceClient)
	}
	for _, server := range serverList {
		if _, ok := grpcServer[server.WebSiteName]; !ok {
			client, err := grpc.Dial(server.ServerUrlList[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return err
			}
			grpcServer[server.WebSiteName] = webSiteGRPC.NewWebSiteServiceClient(client)
		}
	}
	return nil
}
