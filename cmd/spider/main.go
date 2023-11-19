package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"math/rand"
	"os"
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

var (
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
	historyTaskId   int64
	dataPath        string
	config          *Config
	mysqlDatabase   *sql.DB
)

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList(string, chan<- baseStruct.VideoInfo, chan<- baseStruct.TaskClose)
}
type Spider struct {
	interval int64
}

type Config struct {
	DB struct {
		HOST         string `json:"HOST"`
		Port         int    `json:"Port"`
		User         string `json:"User"`
		Password     string `json:"Password"`
		DatabaseName string `json:"DatabaseName"`
	} `json:"DB"`
	Proxy []struct {
		IP    string `json:"IP"`
		HOST  int    `json:"HOST"`
		Token string `json:"Token"`
	} `json:"Proxy"`
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
	fmt.Printf("%v\n", &config)
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

	mysqlDatabase, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.DB.User, config.DB.Password, config.DB.HOST, config.DB.Port, config.DB.DatabaseName))
	if err != nil {
		utils.ErrorLog.Println("数据库连接出错")
		utils.ErrorLog.Println(err.Error())
		return
	}
	//baseStruct.InitDB()
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
	_, err := wheel.AppendOnceFunc(spider.getDynamic, nil, "VideoDynamicSpider", timeWheel.Crontab{ExpiredTime: defaultTicket})
	if err != nil {
		return
	}
	historyTaskId, err = wheel.AppendOnceFunc(spider.getHistory, nil, "VideoHistorySpider", timeWheel.Crontab{ExpiredTime: 10})
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

	dynamicBaseLine := models.GetDynamicBaseline(mysqlDatabase)
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
			website.GetOrCreate(mysqlDatabase)
			author := models.Author{AuthorName: videoInfo.AuthorName, WebSiteId: website.Id, AuthorWebUid: videoInfo.AuthorUuid}
			err = author.GetOrCreate(mysqlDatabase)
			if err != nil {
				utils.ErrorLog.Println(err.Error())
				continue
			}
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
			videoModel.Save(mysqlDatabase)
		case closeInfo = <-closeChan:
			// 删除closeInfo.WebSite的任务
			for index, v := range runWebSite {
				if v == closeInfo.WebSite {
					runWebSite = append(runWebSite[:index], runWebSite[index+1:]...)
					break
				}
			}
			if closeInfo.WebSite == "bilibili" && closeInfo.Code > 0 {
				models.SaveDynamicBaseline(mysqlDatabase, strconv.Itoa(closeInfo.Code))
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

	baseLine := models.GetHistoryBaseLine(mysqlDatabase)
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
	website.GetOrCreate(mysqlDatabase)
	for {
		select {
		case videoInfo := <-VideoHistoryChan:
			author := models.Author{
				AuthorName:   videoInfo.AuthorName,
				WebSiteId:    website.Id,
				AuthorWebUid: videoInfo.AuthorUuid,
				Crawl:        true,
			}
			author.GetOrCreate(mysqlDatabase)
			vi := models.Video{}
			vi.GetByUid(mysqlDatabase, videoInfo.VideoUuid)
			if vi.Id <= 0 {
				vi.CreateTime = videoInfo.PushTime
				vi.Title = videoInfo.Title
				vi.Uuid = videoInfo.VideoUuid
				vi.AuthorId = author.Id
				vi.Save(mysqlDatabase)
			}
			models.VideoHistory{
				WebSiteId: website.Id,
				VideoId:   vi.Id,
				ViewTime:  videoInfo.PushTime,
			}.Save(mysqlDatabase)
		case newestTimestamp := <-VideoHistoryCloseChan:
			models.SaveHistoryBaseLine(mysqlDatabase, strconv.FormatInt(newestTimestamp, 10))
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
	web.GetOrCreate(mysqlDatabase)
	newCollectList := bilibili.Spider.GetCollectList(mysqlDatabase)
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
			author.GetOrCreate(mysqlDatabase)
			vi := models.Video{}
			vi.GetByUid(mysqlDatabase, info.Bvid)
			if vi.Id <= 0 {
				vi.WebSiteId = web.Id
				vi.AuthorId = author.Id
				vi.Title = info.Title
				vi.Desc = info.Intro
				vi.Duration = info.Duration
				vi.Uuid = info.Bvid
				vi.CoverUrl = info.Cover
				vi.UploadTime = time.Unix(info.Ctime, 0)
				vi.Save(mysqlDatabase)
			}
			models.CollectVideo{
				CollectId: collectId,
				VideoId:   vi.Id,
				Mtime:     time.Unix(info.FavTime, 0),
			}.Save(mysqlDatabase)
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
			author.GetOrCreate(mysqlDatabase)
			vi := models.Video{}
			vi.GetByUid(mysqlDatabase, info.Bvid)
			if vi.Id <= 0 {
				vi.WebSiteId = web.Id
				vi.AuthorId = author.Id
				vi.Title = info.Title
				vi.Duration = info.Duration
				vi.Uuid = info.Bvid
				vi.CoverUrl = info.Cover
				vi.Save(mysqlDatabase)
			}
			models.CollectVideo{
				CollectId: collectId,
				VideoId:   vi.Id,
			}.Save(mysqlDatabase)

		}
		time.Sleep(time.Second * 5)
	}

	_, err := wheel.AppendOnceFunc(spider.updateCollectList, nil, "collectListSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + rand.Int63n(100)})
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

	query, err := mysqlDatabase.Query("select cv.collect_id,v.uuid from collect_video cv inner join collect c on c.bv_id = cv.collect_id inner join video v on v.id = cv.video_id where c.`type` = 1 and mtime>'0001-01-01 00:00:00+00:00' order by cv.collect_id,mtime desc")
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
	web.GetOrCreate(mysqlDatabase)
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
				author.GetOrCreate(mysqlDatabase)
				vi := models.Video{}
				vi.GetByUid(mysqlDatabase, info.Bvid)
				if vi.Id <= 0 {
					vi.WebSiteId = web.Id
					vi.AuthorId = author.Id
					vi.Title = info.Title
					vi.Duration = info.Duration
					vi.Uuid = info.Bvid
					vi.CoverUrl = info.Cover
					vi.Save(mysqlDatabase)
				}
				models.CollectVideo{
					CollectId: collectId,
					VideoId:   vi.Id,
				}.Save(mysqlDatabase)
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

	web := models.WebSite{
		WebName: "bilibili",
	}
	web.GetOrCreate(mysqlDatabase)
	resultChan := make(chan bilibili.FollowingUP)
	closeChan := make(chan int64)
	go bilibili.Spider.GetFollowingList(resultChan, closeChan)
	r, err := mysqlDatabase.Query("select author_web_uid from main.author where follow=1")
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
			author.UpdateOrCreate(mysqlDatabase)
		case <-closeChan:
			close = true
		}
		if close {
			break
		}
	}

	_, err = wheel.AppendOnceFunc(spider.updateFollowInfo, nil, "updateFollowInfoSpider", timeWheel.Crontab{ExpiredTime: twelveTicket + rand.Int63n(100)})
	if err != nil {
		utils.ErrorLog.Printf("添加下次运行任务失败：%s\n", err.Error())
		return
	}
}

func deleteRepeatAuthor() {
	//select author_name
	//from author
	//group by author_name
	//having count(*) > 1;

	repeatAuthor, err := mysqlDatabase.Query("select author_name from author group by author_name having count(*) > 1")
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
		_, err = mysqlDatabase.Exec("update video set author_id  = (select id from author where author_name = ? and author.web_site_id = 1),    web_site_id=1 where author_id = (select id from author where author_name = ? and author.web_site_id = 0);", authorName, authorName)
		if err != nil {
			println(err.Error())
			continue
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
