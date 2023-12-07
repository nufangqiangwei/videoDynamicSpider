package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"gorm.io/gorm"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

const (
	defaultTicket = 60 * 5
	oneTicket     = 60 * 60
	sixTime       = oneTicket * 6
	twentyTime    = oneTicket * 20
	twelveTicket  = oneTicket * 12
)

type void struct{}

var (
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
	historyTaskId   int64
	dataPath        string
	config          *Config
	member          void
)

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList(string, chan<- baseStruct.VideoInfo, chan<- baseStruct.TaskClose)
}

type Spider struct {
	interval int64
}

type ProxyInfo struct {
	IP    string `json:"IP"`
	HOST  int    `json:"HOST"`
	Token string `json:"Token"`
}

type Config struct {
	DB struct {
		HOST         string `json:"HOST"`
		Port         int    `json:"Port"`
		User         string `json:"User"`
		Password     string `json:"Password"`
		DatabaseName string `json:"DatabaseName"`
	} `json:"DB"`
	Proxy    []ProxyInfo `json:"Proxy"`
	DataPath string
}

func readConfig() error {
	fileData, err := os.ReadFile("./config.json")
	if err != nil {
		println(err.Error())
		return err
	}
	config = &Config{}
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

	models.InitDB(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DB.User, config.DB.Password, config.DB.HOST, config.DB.Port, config.DB.DatabaseName))
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
	//_, err := wheel.AppendOnceFunc(spider.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: defaultTicket})
	//if err != nil {
	//	return
	//}
	//historyTaskId, err = wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: 10})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendOnceFunc(spider.updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + 120})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendOnceFunc(spider.updateCollectVideoList, nil, "updateCollectVideoSpider", timeWheel.Crontab{ExpiredTime: oneTicket})
	//if err != nil {
	//	return
	//}
	//_, err = wheel.AppendOnceFunc(spider.updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket})
	//if err != nil {
	//	return
	//}
	//wheel.Start()
	spider.getHistory(nil)
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

	dynamicBaseLine := models.GetDynamicBaseline()
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
		err       error
	)
	for {
		select {
		case videoInfo = <-videoResultChan:
			website := models.WebSite{WebName: videoInfo.WebSite}
			err = website.GetOrCreate()
			if err != nil {
				utils.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
				continue
			}
			author := models.Author{AuthorName: videoInfo.AuthorName, WebSiteId: website.Id, AuthorWebUid: videoInfo.AuthorUuid}
			err = author.GetOrCreate()
			if err != nil {
				utils.ErrorLog.Printf("获取作者信息失败：%s\n", err.Error())
				continue
			}
			videoModel := models.Video{
				WebSiteId: website.Id,
				Authors: []models.VideoAuthor{
					{AuthorId: author.Id, Uuid: videoInfo.VideoUuid},
				},
				Title:      videoInfo.Title,
				VideoDesc:  videoInfo.Desc,
				Duration:   videoInfo.Duration,
				Url:        videoInfo.Url,
				Uuid:       videoInfo.VideoUuid,
				CoverUrl:   videoInfo.CoverUrl,
				UploadTime: &videoInfo.PushTime,
			}
			videoModel.Save()

		case closeInfo = <-closeChan:
			// 删除closeInfo.WebSite的任务
			for index, v := range runWebSite {
				if v == closeInfo.WebSite {
					runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
					break
				}
			}
			if closeInfo.WebSite == "bilibili" && closeInfo.Code > 0 {
				models.SaveDynamicBaseline(strconv.Itoa(closeInfo.Code))
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
func (s *Spider) getHistory(interface{}) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			utils.ErrorLog.Println(string(debug.Stack()))
			_, ok := panicErr.(utils.DBFileLock)
			if ok {
				utils.ErrorLog.Println("执行报错，重新添加历史数据爬取")
				var err error
				historyTaskId, err = wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: oneTicket})
				if err != nil {
					utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
				}
				return
			}
			panic(panicErr)
		}
	}()
	utils.Info.Printf("历史任务执行id：%d\n", historyTaskId)

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
	VideoHistoryChan := make(chan baseStruct.VideoInfo)
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
			author := models.Author{
				AuthorName:   videoInfo.AuthorName,
				WebSiteId:    website.Id,
				AuthorWebUid: videoInfo.AuthorUuid,
				Crawl:        true,
			}
			err = author.GetOrCreate()
			if err != nil {
				utils.ErrorLog.Printf("获取作者信息失败：%s\n", err.Error())
				continue
			}
			vi := models.Video{
				CreateTime: videoInfo.PushTime,
				Title:      videoInfo.Title,
				Uuid:       videoInfo.VideoUuid,
				Authors: []models.VideoAuthor{
					{AuthorId: author.Id, Uuid: videoInfo.VideoUuid},
				},
			}
			vi.Save()

			models.VideoHistory{
				WebSiteId: website.Id,
				VideoId:   vi.Id,
				ViewTime:  videoInfo.PushTime,
				WebUUID:   videoInfo.VideoUuid,
			}.Save()
		case newestTimestamp := <-VideoHistoryCloseChan:
			models.SaveHistoryBaseLine(strconv.FormatInt(newestTimestamp, 10))
			historyTaskId, err = wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: historyRunTime()})
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

	web := models.WebSite{
		WebName: "bilibili",
	}
	err := web.GetOrCreate()
	if err != nil {
		utils.ErrorLog.Printf("获取网站信息失败：%s\n", err.Error())
		return
	}
	newCollectList := bilibili.Spider.GetCollectList()
	for _, collectId := range newCollectList.Collect {
		saveBilibiliCollectAllVideo(bilibili.Spider.GetCollectAllVideo(collectId, 0), web.Id, collectId)
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
					{AuthorId: author.Id, Uuid: info.Bvid},
				},
			}
			vi.Save()
			models.CollectVideo{
				CollectId: collectId,
				VideoId:   vi.Id,
			}.Save()

		}
		time.Sleep(time.Second * 5)
	}

	_, err = wheel.AppendOnceFunc(spider.updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + rand.Int63n(100)})
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
				author := models.Author{
					WebSiteId:    web.Id,
					AuthorWebUid: strconv.Itoa(info.Upper.Mid),
					AuthorName:   info.Upper.Name,
					Follow:       false,
					Crawl:        true,
				}
				author.GetOrCreate()
				vi := models.Video{
					WebSiteId: web.Id,
					Title:     info.Title,
					Duration:  info.Duration,
					Uuid:      info.Bvid,
					CoverUrl:  info.Cover,
					Authors: []models.VideoAuthor{
						{AuthorId: author.Id, Uuid: info.BvId},
					},
				}
				vi.Save()
				models.CollectVideo{
					CollectId: collectId,
					VideoId:   vi.Id,
				}.Save()
			}
		}
	}

	_, err := wheel.AppendOnceFunc(spider.updateCollectVideoList, nil, "updateCollectVideoSpider", timeWheel.Crontab{ExpiredTime: arrangeRunTime(twelveTicket, sixTime, twentyTime)})
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
			fmt.Printf("%v\n", upInfo)
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
				utils.ErrorLog.Printf("更新作者信息失败：%s\n", err.Error())
				continue
			}
		}
	}
	_, err = wheel.AppendOnceFunc(spider.updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + rand.Int63n(100)})
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

func saveAuthorAllVideo(db *sql.DB, response bilibili.VideoListPageResponse, webSiteId int64) []string {
	var (
		authorVideoUUID map[string]string
		videoUUID       string
		ok              bool
		result          []string
		authorId        int64
		err             error
	)
	if len(response.Data.List.Vlist) > 0 {
		err = db.QueryRow("select id from main.author where author_web_uid=?", response.Data.List.Vlist[0].Mid).Scan(&authorId)
		if err != nil {
			utils.ErrorLog.Printf(err.Error())
			return nil
		}
	} else {
		return []string{}
	}
	uuidQuery, err := db.Query("select uuid from video where author_id=?", authorId)
	if err != nil {
		utils.ErrorLog.Printf("saveAuthorAllVideo方法查询已有存有视频信息出错")
		return nil
	}

	for uuidQuery.Next() {
		err = uuidQuery.Scan(&videoUUID)
		if err == nil {
			authorVideoUUID[videoUUID] = ""
		}
	}
	uuidQuery.Close()

	for _, videoInfo := range response.Data.List.Vlist {
		if _, ok = authorVideoUUID[videoInfo.Bvid]; ok {
			continue
		}
		result = append(result,
			fmt.Sprintf("(%d, %d,'%s','%s',%d,'%s','%s','%s')", webSiteId, authorId, videoInfo.Title, strings.Replace(videoInfo.Description, " ", "", -1),
				bilibili.HourAndMinutesAndSecondsToSeconds(videoInfo.Length), videoInfo.Bvid, videoInfo.Pic, time.Unix(int64(videoInfo.Created), 0).Format("2006-01-02 15:04:05-07:00")))

	}
	return result
}

// 更新正在代理中正在运行的任务状态
func updateProxySpiderTaskStatus() {
	// 获取正在运行的任务
	runTaskList := make([]models.ProxySpiderTask, 0)
	tx := models.GormDB.Where("status in ?", []int{1, 2}).Find(runTaskList)
	if tx.Error != nil {
		utils.ErrorLog.Println("获取正在运行的任务失败")
		return
	}
	// 向代理的getTaskStatus这个路径发起get请求获取任务状态
	var responseData struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Md5    string `json:"md5"`
	}
	for _, proxyInfo := range runTaskList {
		request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/getTaskStatus", proxyInfo.SpiderIp), nil)
		q := request.URL.Query()
		q.Add("taskType", proxyInfo.TaskType)
		q.Add("taskId", proxyInfo.TaskId)
		request.URL.RawQuery = q.Encode()
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			utils.ErrorLog.Printf("获取任务状态失败：%s\n", err.Error())
			continue
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			utils.ErrorLog.Printf("读取响应失败：%s\n", err.Error())
			continue
		}
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			utils.ErrorLog.Printf("解析响应失败：%s\n", err.Error())
			continue
		}
		models.GormDB.Model(&proxyInfo).Update("status", responseData.Status)
		if responseData.Status == 2 {
			models.GormDB.Model(&proxyInfo).Update("result_file_md5", responseData.Md5)
			// 任务完成
			go downloadTaskResult(proxyInfo.SpiderIp, proxyInfo.TaskType, proxyInfo.TaskId, responseData.Md5)
		}
		response.Body.Close()
	}
	// 更新任务状态
}

func downloadTaskResult(ip, taskType, taskId, fileMd5 string) {
	// 检查文件是否已经下载了
	if m, _ := utils.GetFileMd5(path.Join(dataPath, taskType, taskId, "result")); m == fileMd5 {
		return
	}
	// 直接去下载文件,服务方使用nginx做文件服务器，下载根路径是 taskResult
	request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/taskResult/%s/%s", ip, taskType, taskId), nil)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		utils.ErrorLog.Printf("获取任务结果失败：%s\n", err.Error())
		return
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		utils.ErrorLog.Printf("读取响应失败：%s\n", err.Error())
		return
	}
	// 将文件写入到本地
	err = ioutil.WriteFile(path.Join(dataPath, taskType, taskId, "result"), data, fs.ModePerm)
	if err != nil {
		utils.ErrorLog.Printf("写入文件失败：%s\n", err.Error())
		return
	}
}

func saveBilibiliCollectAllVideo(response []bilibili.CollectVideoDetailInfo, webSiteId, collectId int64) {
	for _, info := range response {
		author := models.Author{
			WebSiteId:    webSiteId,
			AuthorWebUid: strconv.Itoa(info.Upper.Mid),
			AuthorName:   info.Upper.Name,
			Avatar:       info.Upper.Face,
			Follow:       false,
			Crawl:        true,
		}
		err := author.GetOrCreate()
		if err != nil {
			utils.ErrorLog.Printf("获取作者信息失败：%s\n", err.Error())
			continue
		}
		uploadTime := time.Unix(info.Ctime, 0)
		vi := models.Video{
			WebSiteId: webSiteId,
			Authors: []models.VideoAuthor{
				{AuthorId: author.Id, Uuid: info.BvId},
			},
			Title:      info.Title,
			VideoDesc:  info.Intro,
			Duration:   info.Duration,
			Uuid:       info.Bvid,
			CoverUrl:   info.Cover,
			UploadTime: &uploadTime,
		}
		vi.Save()
		models.CollectVideo{
			CollectId: collectId,
			VideoId:   vi.Id,
			Mtime:     time.Unix(info.FavTime, 0),
		}.Save()
	}
}

func saveBilibiliAuthorVideoList(response bilibili.VideoListPageResponse, WebSiteId, authorId int64, authorVideoUUIDMap map[string]struct{}) {
	if len(response.Data.List.Vlist) == 0 {
		return
	}
	if authorId == 0 {
		authorMid := response.Data.List.Vlist[0].Mid
		// 查询这个作者的id
		models.GormDB.Table("author").
			Select("id").
			Where("author_web_uid = ?", authorMid).
			Find(&authorId)
	}
	if len(authorVideoUUIDMap) == 0 {
		var authorVideoUUIDList []string
		// 查询这个作者本地保存的视频信息
		models.GormDB.Table("video v").
			Select("v.uuid").
			Where("a.author_id = ?", authorId).
			Find(&authorVideoUUIDList)
		for _, videoUuid := range authorVideoUUIDList {
			authorVideoUUIDMap[videoUuid] = member
		}
	}
	var (
		ok          bool
		insertVideo []*models.Video
	)
	for _, videoInfo := range response.Data.List.Vlist {
		_, ok = authorVideoUUIDMap[videoInfo.Bvid]
		if !ok {
			createdTime := time.Unix(videoInfo.Created, 0)
			// 保存视频信息
			vv := models.Video{
				WebSiteId: WebSiteId,
				Authors: []models.VideoAuthor{
					{AuthorId: authorId, Uuid: videoInfo.Bvid},
				},
				Title:      videoInfo.Title,
				VideoDesc:  videoInfo.Description,
				Duration:   bilibili.HourAndMinutesAndSecondsToSeconds(videoInfo.Length),
				Uuid:       videoInfo.Bvid,
				Url:        "",
				CoverUrl:   videoInfo.Pic,
				UploadTime: &createdTime,
				CreateTime: time.Now(),
			}
			insertVideo = append(insertVideo, &vv)
			authorVideoUUIDMap[videoInfo.Bvid] = member
		}
	}
	models.GormDB.Create(insertVideo)
}

func updateBilibiliVideoDetailInfo(response bilibili.VideoDetailResponse, WebSiteId int64) {
	video := models.Video{}
	var tx *gorm.DB
	tx = models.GormDB.Where("uuid = ?", response.Data.View.Bvid).Preload("Authors").Preload("Tag").
		Limit(1).Find(&video)
	if tx.Error != nil {
		utils.ErrorLog.Printf("获取视频信息失败：%s\n", tx.Error.Error())
		return
	}
	if video.Id == 0 {
		// 视频不存在，video表中创建这条视频数据
		uploadTime := time.Unix(response.Data.View.Ctime, 0)
		video = models.Video{
			WebSiteId:  WebSiteId,
			Title:      response.Data.View.Title,
			Uuid:       response.Data.View.Bvid,
			CoverUrl:   response.Data.View.Pic,
			VideoDesc:  response.Data.View.Desc,
			CreateTime: time.Now(),
			UploadTime: &uploadTime,
			Duration:   response.Data.View.Duration,
		}
		models.GormDB.Create(&video)
	}
	// 更新视频信息
	video.View = response.Data.View.Stat.View
	video.Danmaku = response.Data.View.Stat.Danmaku
	video.Reply = response.Data.View.Stat.Reply
	video.Favorite = response.Data.View.Stat.Favorite
	video.Coin = response.Data.View.Stat.Coin
	video.Share = response.Data.View.Stat.Share
	video.NowRank = response.Data.View.Stat.NowRank
	video.HisRank = response.Data.View.Stat.HisRank
	video.Like = response.Data.View.Stat.Like
	video.Dislike = response.Data.View.Stat.Dislike
	video.Evaluation = response.Data.View.Stat.Evaluation
	models.GormDB.Model(&video).Updates(map[string]interface{}{
		"View":       response.Data.View.Stat.View,
		"Danmaku":    response.Data.View.Stat.Danmaku,
		"Reply":      response.Data.View.Stat.Reply,
		"Favorite":   response.Data.View.Stat.Favorite,
		"Coin":       response.Data.View.Stat.Coin,
		"Share":      response.Data.View.Stat.Share,
		"NowRank":    response.Data.View.Stat.NowRank,
		"HisRank":    response.Data.View.Stat.HisRank,
		"Like":       response.Data.View.Stat.Like,
		"Dislike":    response.Data.View.Stat.Dislike,
		"Evaluation": response.Data.View.Stat.Evaluation,
	})
	// 更新作者和协作者信息
	// 查询作者信息
	DatabaseAuthorInfo := []models.Author{}
	authorIdList := []int64{}
	for _, a := range video.Authors {
		authorIdList = append(authorIdList, a.AuthorId)
	}
	models.GormDB.Where("id in ?", authorIdList).Find(&DatabaseAuthorInfo)
	// models.VideoAuthor 和 response.Data.View.Staff 两边信息做对比，models.VideoAuthor缺少的就添加，models.Author缺少的就添加
	authorHave := false
	for _, b := range response.Data.View.Staff {
		for _, a := range DatabaseAuthorInfo {
			if a.AuthorWebUid == strconv.Itoa(b.Mid) {
				authorHave = true
				break
			}
		}
		if !authorHave {
			// 查询这个作者在Author表中是否存在
			author := models.Author{}
			models.GormDB.Where("author_web_uid = ?", strconv.Itoa(b.Mid)).Find(&author)
			if author.Id == 0 {
				// 作者不存在，数据库中添加作者信息
				author = models.Author{
					AuthorName:   b.Name,
					WebSiteId:    WebSiteId,
					AuthorWebUid: strconv.Itoa(b.Mid),
					Avatar:       b.Face,
					FollowNumber: b.Follower,
				}
				models.GormDB.Create(&author)
			}
			va := models.VideoAuthor{
				Uuid:       response.Data.View.Bvid,
				VideoId:    video.Id,
				AuthorId:   author.Id,
				Contribute: b.Title,
			}
			models.GormDB.Create(&va)
		}
	}

	// 更新视频标签信息
	var tagHave bool
	for _, v := range response.Data.Tags {
		// 循环video.Tag，如果有这个标签，就标记已存在,并且在videoTag中删除这个标签
		tagHave = false
		for index, tag := range video.Tag {
			if tag.Id == v.TagId {
				tagHave = true
				video.Tag = append(video.Tag[:index], video.Tag[index+1:]...)
				break
			}
		}

		if !tagHave {
			tag := models.Tag{}
			models.GormDB.Find(&tag, "id=?", v.TagId)
			if tag.Name == "" {
				tag.Id = v.TagId
				tag.Name = v.TagName
				models.GormDB.Create(&tag)
			}
			videoTag := models.VideoTag{
				VideoId: video.Id,
				TagId:   v.TagId,
			}
			models.GormDB.Create(&videoTag)
		}

	}

}
