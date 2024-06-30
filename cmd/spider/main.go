package main

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	"videoDynamicAcquisition/grpcDiscovery/redis"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
	webSiteGRPC "videoDynamicAcquisition/proto"
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

	})
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
	_, err = wheel.AppendCycleFunc(updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twentyTime})
	if err != nil {
		return
	}
	_, err = wheel.AppendCycleFunc(updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twentyTime + rand.Int63n(100)*50})
	if err != nil {
		return
	}
	_, err = wheel.AppendCycleFunc(textUpdateVideoInfo, nil, "textUpdateVideoInfoSpider", timeWheel.Crontab{ExpiredTime: oneTicket})
	if err != nil {
		return
	}

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
			lastUpdateTime := models.GetDynamicBaseline(userCookie.GetDBPrimaryKeyId())
			cookiesMap := userCookie.GetCookiesDictToLowerKey()
			cookiesMap = checkOutWebSiteCookies(websiteServerName, cookiesMap)
			go func(websiteServerName, userName, lastUpdateTime string, webSiteId, userId int64, cookieMap map[string]string) {
				cookieMap["requestUserName"] = userName
				// 粗暴的设置超时时间为5分钟
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				res, err := server.GetUserFollowUpdate(ctx, &webSiteGRPC.UserInfo{
					Cookies:        cookieMap,
					LastUpdateTime: lastUpdateTime,
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
			}(websiteServerName, userName, lastUpdateTime, userCookie.GetWebSiteId(), userCookie.GetDBPrimaryKeyId(), cookiesMap)
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
					Authors: []*models.VideoAuthor{
						{Contribute: "UP主", AuthorUUID: videoInfoResponse.Authors[0].Uid},
					},
					StructAuthor: []*models.Author{
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
	if len(grpcServer) == 0 {
		println("没有可用的GRPC服务")
		return
	}
	historyResult = make(chan *webSiteGRPC.VideoInfoResponse, 50)
	for websiteServerName, server := range grpcServer {
		for userName, userCookie := range cookies.GetWebSiteUser(websiteServerName) {
			lastUpdateTime := models.GetHistoryBaseline(userCookie.GetDBPrimaryKeyId())
			cookiesMap := userCookie.GetCookiesDictToLowerKey()
			cookiesMap = checkOutWebSiteCookies(websiteServerName, cookiesMap)
			go func(websiteServerName, userName, lastUpdateTime string, webSiteId, userId int64, cookieMap map[string]string) {
				cookieMap["requestUserName"] = userName
				// 粗暴的设置超时时间为5分钟
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				res, err := server.GetUserViewHistory(ctx, &webSiteGRPC.UserInfo{
					Cookies:         cookieMap,
					LastHistoryTime: lastUpdateTime,
				})
				if err != nil {
					log.ErrorLog.Printf("GRPC服务端%s获取历史记录失败:%s\n", userName, err.Error())
					historyResult <- &webSiteGRPC.VideoInfoResponse{
						ErrorCode:       500,
						ErrorMsg:        err.Error(),
						WebSiteName:     websiteServerName,
						RequestUserName: userName,
						WebSiteId:       webSiteId,
						RequestUserId:   userId,
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
							WebSiteId:       webSiteId,
							RequestUserId:   userId,
						}
					}
					vvi.WebSiteName = websiteServerName
					vvi.RequestUserName = userName
					vvi.WebSiteId = webSiteId
					vvi.RequestUserId = userId

					historyResult <- vvi
					if vvi.ErrorCode != 0 {
						if vvi.ErrorCode != 200 {
							log.ErrorLog.Printf("%s获取历史状态错误:%s\n", userName, vvi.ErrorMsg)
						}
						break
					}
				}
			}(websiteServerName, userName, lastUpdateTime, userCookie.GetWebSiteId(), userCookie.GetDBPrimaryKeyId(), cookiesMap)
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
				Authors: []*models.VideoAuthor{
					{
						Contribute: "UP主",
						AuthorUUID: vi.Authors[0].Uid,
						Uuid:       vi.Uid,
					},
				},
				StructAuthor: []*models.Author{
					{
						AuthorName:   vi.Authors[0].Name,
						AuthorWebUid: vi.Authors[0].Uid,
						Avatar:       vi.Authors[0].Avatar,
					},
				},
				ViewHistory: []*models.VideoHistory{
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
				break
			}
		}
	}
}

// 更新收藏夹列表，喝订阅的合集列表，新创建的同步视频数据
func updateCollectList(interface{}) {
	var runWebSite = make([]string, 0)
	historyResult := make(chan *webSiteGRPC.CollectionInfo, 50)
	for websiteServerName, server := range grpcServer {
		for userName, userCookie := range cookies.GetWebSiteUser(websiteServerName) {
			cookiesMap := userCookie.GetCookiesDictToLowerKey()
			cookiesMap = checkOutWebSiteCookies(websiteServerName, cookiesMap)
			cookiesMap["requestUserName"] = userName
			go func(websiteServerName, userName string, webSiteId, userId int64, cookieMap map[string]string) {
				// 粗暴的设置超时时间为5分钟
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				res, err := server.GetUserCollectionList(ctx, &webSiteGRPC.CollectionInfoRequest{
					User: &webSiteGRPC.UserInfo{
						Cookies: cookieMap,
					},
				})
				if err != nil {
					log.ErrorLog.Printf("GRPC服务端%s获取收藏夹信息失败:%s\n", userName, err.Error())
					historyResult <- &webSiteGRPC.CollectionInfo{
						ErrorCode:       500,
						ErrorMsg:        err.Error(),
						WebSiteName:     websiteServerName,
						RequestUserName: userName,
						WebSiteId:       webSiteId,
						RequestUserId:   userId,
					}
					return
				}
				for {
					vvi, err := res.Recv()
					if err == io.EOF {
						return
					}
					if err != nil {
						log.ErrorLog.Printf("GRPC服务端%s获取收藏夹信息流通道错误:%s\n", userName, err.Error())
						vvi = &webSiteGRPC.CollectionInfo{
							ErrorCode:       500,
							ErrorMsg:        err.Error(),
							WebSiteName:     websiteServerName,
							RequestUserName: userName,
							WebSiteId:       webSiteId,
							RequestUserId:   userId,
						}
					}
					vvi.WebSiteName = websiteServerName
					vvi.RequestUserName = userName
					vvi.WebSiteId = webSiteId
					vvi.RequestUserId = userId

					historyResult <- vvi
					if vvi.ErrorCode != 0 && vvi.ErrorCode != -404 {
						if vvi.ErrorCode != 200 {
							log.ErrorLog.Printf("%s获取收藏夹信息错误:%s\n", userName, vvi.ErrorMsg)
						}
						break
					}
				}
			}(websiteServerName, userName, userCookie.GetWebSiteId(), userCookie.GetDBPrimaryKeyId(), cookiesMap)
			runWebSite = append(runWebSite, fmt.Sprintf("%s-%s", websiteServerName, userName))
			time.Sleep(time.Second)
		}
	}

	for vi := range historyResult {
		if vi.ErrorCode == 0 {
			handleUserCollectList(vi)
		} else {
			if vi.ErrorCode == -404 {
				models.GormDB.Exec(`update collect set is_invalid = true where bv_id = ?`, vi.CollectionId)
				continue
			}
			userInfo := fmt.Sprintf("%s-%s", vi.WebSiteName, vi.RequestUserName)
			for index, v := range runWebSite {
				if v == userInfo {
					runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
					break
				}
			}
			if len(runWebSite) == 0 {
				break
			}
		}
	}
}

// 同步关注信息
func updateFollowInfo(interface{}) {
	var (
		authorInfoResult chan *webSiteGRPC.AuthorInfoResponse
		runWebSite       = make([]string, 0)
	)
	authorInfoResult = make(chan *webSiteGRPC.AuthorInfoResponse)
	println("开始同步关注信息")
	fmt.Printf("%v\n", grpcServer)
	if len(grpcServer) == 0 {
		println("没有可用的GRPC服务")
		return
	}
	for websiteServerName, server := range grpcServer {
		println(websiteServerName)
		for userName, userCookie := range cookies.GetWebSiteUser(websiteServerName) {
			println(userName)
			cookiesMap := userCookie.GetCookiesDictToLowerKey()
			cookiesMap = checkOutWebSiteCookies(websiteServerName, cookiesMap)
			go func(websiteServerName, userName string, webSiteId, userId int64, cookieMap map[string]string) {
				cookieMap["requestUserName"] = userName
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				res, err := server.GetUserFollowList(ctx, &webSiteGRPC.UserInfo{Cookies: cookieMap})
				if err != nil {
					log.ErrorLog.Printf("GRPC服务端%s获取用户关注失败:%s\n", userName, err.Error())
					authorInfoResult <- &webSiteGRPC.AuthorInfoResponse{
						ErrorCode:       500,
						ErrorMsg:        err.Error(),
						WebSiteName:     websiteServerName,
						RequestUserName: userName,
						WebSiteId:       webSiteId,
						RequestUserId:   userId,
					}
					return
				}
				for {
					vvi, err := res.Recv()
					if err == io.EOF {
						return
					}
					if err != nil {
						log.ErrorLog.Printf("GRPC服务端%s获取用户关注通道错误:%s\n", userName, err.Error())
						vvi = &webSiteGRPC.AuthorInfoResponse{
							ErrorCode: 500,
							ErrorMsg:  err.Error(),
						}
					}
					vvi.WebSiteName = websiteServerName
					vvi.RequestUserName = userName
					vvi.WebSiteId = webSiteId
					vvi.RequestUserId = userId

					authorInfoResult <- vvi
					if vvi.ErrorCode != 0 {
						if vvi.ErrorCode != 200 {
							log.ErrorLog.Printf("%s获取历史状态错误:%s\n", userName, vvi.ErrorMsg)
						}
						break
					}
				}
			}(websiteServerName, userName, userCookie.GetWebSiteId(), userCookie.GetDBPrimaryKeyId(), cookiesMap)
			runWebSite = append(runWebSite, fmt.Sprintf("%s-%s", websiteServerName, userName))
		}
	}

	var grpcResult = make(map[int64][]*webSiteGRPC.AuthorInfoResponse)
	for authorInfo := range authorInfoResult {
		if authorInfo.ErrorCode == 0 {
			_, ok := grpcResult[authorInfo.RequestUserId]
			if !ok {
				grpcResult[authorInfo.RequestUserId] = make([]*webSiteGRPC.AuthorInfoResponse, 0)
			}
			grpcResult[authorInfo.RequestUserId] = append(grpcResult[authorInfo.RequestUserId], authorInfo)
		} else {
			userInfo := fmt.Sprintf("%s-%s", authorInfo.WebSiteName, authorInfo.RequestUserName)
			fmt.Printf("%d %s %s\n", authorInfo.ErrorCode, authorInfo.ErrorMsg, userInfo)
			for index, v := range runWebSite {
				if v == userInfo {
					runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
					break
				}
			}
			if len(runWebSite) == 0 {
				break
			}
		}

	}

	for userId, authorInfo := range grpcResult {
		userFollow := models.GetUserFollowList(userId)
		// 根据 webSiteGRPC.AuthorInfoResponse.Uid，models.FollowRelation.AuthorWebUid 找出authorInfo，userFollow两个列表的差集
		newAuthorUid := make([]string, 0)
		for _, i := range authorInfo {
			newAuthorUid = append(newAuthorUid, i.Uid)
		}
		oldAuthorUid := make([]string, 0)
		for _, i := range userFollow {
			oldAuthorUid = append(oldAuthorUid, i.AuthorWebUid)
		}
		// 删除的关注作者
		for _, delAuthor := range utils.ArrayDifference(oldAuthorUid, newAuthorUid) {
			models.DeleteAuthorFollowByUid(userId, delAuthor)
		}
		// 新增的关注作者
		for _, addAuthor := range utils.ArrayDifference(newAuthorUid, oldAuthorUid) {
			author := models.Author{}
			author.GetByUid(addAuthor)
			var newAuthorInfo *webSiteGRPC.AuthorInfoResponse
			for _, i := range authorInfo {
				if i.Uid == addAuthor {
					newAuthorInfo = i
					break
				}
			}
			if newAuthorInfo == nil {
				continue
			}
			if author.Id == 0 {
				author.AuthorName = newAuthorInfo.Name
				author.AuthorWebUid = newAuthorInfo.Uid
				author.Avatar = newAuthorInfo.Avatar
				author.AuthorDesc = newAuthorInfo.Desc
				author.FollowNumber = newAuthorInfo.FollowNumber
				author.UpdateOrCreate()
			}
			followTime := time.Unix(newAuthorInfo.FollowTime, 0)
			models.GormDB.Create(&models.Follow{
				Id:         0,
				WebSiteId:  authorInfo[0].WebSiteId,
				AuthorId:   author.Id,
				UserId:     userId,
				FollowTime: &followTime,
				Deteled:    false,
			})
		}
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
				var authors []*models.VideoAuthor
				var structAuthor []*models.Author
				// 是否是联合投稿 0：否 1：是
				if videoInfo.IsUnionVideo == 0 {
					authors = append(authors, &models.VideoAuthor{Uuid: videoInfo.Bvid, AuthorUUID: author.AuthorWebUid})
					structAuthor = append(structAuthor, &models.Author{
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
			DatabaseAuthorInfo := []*models.Author{
				{
					WebSiteId:    videoInfo.WebSiteId,
					AuthorName:   response.Data.View.Owner.Name,
					AuthorWebUid: strconv.FormatInt(response.Data.View.Owner.Mid, 10),
					Avatar:       response.Data.View.Owner.Face,
				},
			}
			if len(response.Data.View.Staff) > 0 {
				for _, staff := range response.Data.View.Staff {
					DatabaseAuthorInfo = append(DatabaseAuthorInfo, &models.Author{
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
	var authorList []models.Author
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
					Authors: []*models.VideoAuthor{
						{Contribute: "UP主", AuthorUUID: strconv.Itoa(dynmaicInfo.Modules.ModuleAuthor.Mid)},
					},
					StructAuthor: []*models.Author{
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

func textUpdateVideoInfo(interface{}) {
	var videoList []*models.Video
	models.GormDB.Model(&models.Video{}).Where("upload_time > now() - interval 30 day").Scan(&videoList)
	println("待更新", len(videoList), "个视频")
	if len(videoList) > 0 {
		getVideoDetail(videoList)
	}

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

func checkOutWebSiteCookies(websiteName string, cookiesMap map[string]string) map[string]string {
	switch websiteName {
	case "bilibili":
		return bilibiliCheckOutCookies(cookiesMap)
	}
	return cookiesMap
}
func bilibiliCheckOutCookies(cookiesMap map[string]string) map[string]string {
	cookieKeyList := []string{
		"sessdata",
		"bili_jct",
		"buvid3",
		"dedeuserid",
		"ac_time_value",
	}
	result := make(map[string]string)
	for k, v := range cookiesMap {
		for _, key := range cookieKeyList {
			if strings.Contains(k, key) {
				result[key] = v
			}
		}
	}
	return result
}

func updateVideoDetailInfo(videoGrpcDetail *webSiteGRPC.VideoInfoResponse) {
	transactions := models.GormDB.Begin()
	video := models.GetVideoFullData(transactions, videoGrpcDetail.WebSiteId, videoGrpcDetail.Uid)
	if video.Id <= 0 {
		video = videoGrpcDetail.ToVideoModel()
		err := transactions.Create(video).Error
		if err != nil {
			println(err.Error())
			return
		}
	}
	videoTbaleUpdate := map[string]interface{}{}
	if video.VideoDesc != videoGrpcDetail.Desc {
		videoTbaleUpdate["video_desc"] = videoGrpcDetail.Desc
	}
	if video.Duration == 0 {
		videoTbaleUpdate["duration"] = videoGrpcDetail.Duration
	}
	if video.CoverUrl != videoGrpcDetail.Cover {
		videoTbaleUpdate["cover_url"] = videoGrpcDetail.Cover
	}
	if video.UploadTime == nil || video.UploadTime.IsZero() {
		t := time.Unix(videoGrpcDetail.UpdateTime, 0)
		videoTbaleUpdate["upload_time"] = &t
	}
	if len(videoTbaleUpdate) > 0 {
		err := transactions.Model(&video).Updates(videoTbaleUpdate).Error
		if err != nil {
			println(err.Error())
			return
		}
	}
	// 更新作者信息
	authorArrayUnique := utils.ArrayUnique{
		LeftArray:  videoGrpcDetail.Authors,
		RightArray: video.StructAuthor,
	}
	// 需要更新的作者参与信息
	for _, info := range authorArrayUnique.GetIntersection() {
		remoteInfo := info.Left.(*webSiteGRPC.AuthorInfoResponse)
		localInfo := info.Right.(*models.Author)
		for _, va := range video.Authors {
			if va.AuthorId == localInfo.Id {
				if va.Contribute != remoteInfo.Author {
					transactions.Model(&va).Update("contribute", remoteInfo.Author)
				}
				break
			}
		}
	}
	// 需要新增的作者
	for _, info := range authorArrayUnique.GetInLeftObject() {
		authorInfo := info.(*webSiteGRPC.AuthorInfoResponse)
		author := authorInfo.ToModel()
		author.GetOrCreate()
		authorVideo := models.VideoAuthor{
			Uuid:       videoGrpcDetail.Uid,
			VideoId:    video.Id,
			AuthorId:   author.Id,
			Contribute: authorInfo.Author,
			AuthorUUID: authorInfo.Uid,
		}
		transactions.Create(&authorVideo)
	}

	// 更新标签信息
	tagArrayUnique := utils.ArrayUnique{
		LeftArray:  videoGrpcDetail.Tags,
		RightArray: video.StructTag,
	}
	// 需要更新的标签参与信息
	for _, info := range tagArrayUnique.GetIntersection() {
		remoteInfo := info.Left.(*webSiteGRPC.TagInfoResponse)
		localInfo := info.Right.(*models.Tag)
		for _, va := range video.Tag {
			if va.TagId == localInfo.Id {
				if va.TagId != localInfo.Id {
					transactions.Model(&va).Update("tag_name", remoteInfo.Name)
				}
				break
			}
		}
	}
	// 需要新增的标签
	for _, info := range tagArrayUnique.GetInLeftObject() {
		tagInfo := info.(*webSiteGRPC.TagInfoResponse)
		tag := tagInfo.ToModel()
		tag.GetOrCreate()
		tagVideo := models.VideoTag{
			VideoId: video.Id,
			TagId:   tag.Id,
		}
		transactions.Create(&tagVideo)
	}

	// 更新视频分区信息
	if videoGrpcDetail.Classify != nil {
		if video.Classify == nil {
			classify := videoGrpcDetail.Classify
			classifyInfo := models.Classify{
				Id:   classify.Id,
				Name: classify.Name,
			}
			transactions.Create(&classifyInfo)
		} else if video.Classify.Name != videoGrpcDetail.Classify.Name {
			transactions.Model(&video).Update("classify_name", videoGrpcDetail.Classify.Name)
		}
	}

	// 插入播放数据
	vp := models.VideoPlayData{
		VideoId:    video.Id,
		View:       videoGrpcDetail.ViewNumber,
		Danmaku:    videoGrpcDetail.Danmaku,
		Reply:      videoGrpcDetail.Reply,
		Favorite:   videoGrpcDetail.Favorite,
		Coin:       videoGrpcDetail.Coin,
		Share:      videoGrpcDetail.Share,
		Like:       videoGrpcDetail.Like,
		Dislike:    videoGrpcDetail.Dislike,
		NowRank:    videoGrpcDetail.NowRank,
		HisRank:    videoGrpcDetail.HisRank,
		Evaluation: videoGrpcDetail.Evaluation,
		CreateTime: time.Now(),
	}
	transactions.Create(&vp)
	transactions.Commit()
}

func getVideoDetail(videoList []*models.Video) {
	var webSiteList []models.WebSite
	err := models.GormDB.Model(&models.WebSite{}).Find(&webSiteList).Error
	if err != nil {
		println(err.Error())
		return
	}
	webSiteMap := map[int64]string{}
	for _, webSiteInfo := range webSiteList {
		webSiteMap[webSiteInfo.Id] = webSiteInfo.WebName
	}

	var webSiteIdList []int64
	groupByWebSiteId := make(map[int64][]*models.Video)
	for _, video := range videoList {
		webSiteIdList = append(webSiteIdList, video.WebSiteId)
		vList, ok := groupByWebSiteId[video.WebSiteId]
		if ok {
			groupByWebSiteId[video.WebSiteId] = []*models.Video{}
		}
		vList = append(vList, video)
	}
	for _, webSiteId := range webSiteIdList {
		go func(requestWebSiteId int64) {
			webSiteName := webSiteMap[requestWebSiteId]
			server, ok := grpcServer[webSiteName]
			if !ok {
				println(webSiteName, " 服务不存在")
				return
			}
			stream, err := server.GetVideoDetail(context.Background())
			if err != nil {
				println(webSiteName, "拨号错误")
				println(err.Error())
				return
			}
			go func() {
				for {
					videoDetailResponse, err := stream.Recv()
					if err != nil {
						println(webSiteName, "接收错误")
						println(err.Error())
						return
					}
					videoDetail := videoDetailResponse.GetVideoDetail()
					videoDetail.WebSiteId = requestWebSiteId
					videoDetail.WebSiteName = webSiteName
					updateVideoDetailInfo(videoDetail)
					for _, videoDetail = range videoDetailResponse.GetRecommendVideo() {
						videoDetail.WebSiteId = requestWebSiteId
						videoDetail.WebSiteName = webSiteName
						updateVideoDetailInfo(videoDetail)
					}
				}
			}()
			for _, video := range groupByWebSiteId[requestWebSiteId] {
				err = stream.Send(&webSiteGRPC.GetVideoListRequest{
					VideoIdList: video.Uuid,
				})
				time.Sleep(time.Millisecond * 100)
			}

		}(webSiteId)
		time.Sleep(time.Second * 1)
	}

}
