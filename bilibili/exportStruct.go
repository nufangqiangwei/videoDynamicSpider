package bilibili

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

type BiliSpider struct {
}

const (
	defaultUpdateNumber = 100
	webSiteName         = "bilibili"
	selfUserMid         = "DedeUserID" // 从cookies中提取用户的mid
)

var (
	Spider             = BiliSpider{}
	wbiSignObj         = wbiSign{lastUpdateTime: time.Now()}
	dynamicBaseLineMap map[string]int64
	historyBaseLineMap map[string]int64
	webSiteId          int64
	historyRequestSave map[string]utils.WriteFile
)

func (s BiliSpider) Init(dynamicBaseLine map[string]int64, historyBaseLine map[string]int64, webSiteTableId int64) {
	dynamicBaseLineMap = dynamicBaseLine
	historyBaseLineMap = historyBaseLine
	fmt.Printf("bilibili初始化成功,\n网站ID:%d,\n动态基线:%v,\n历史基线:%v,\n", webSiteTableId, dynamicBaseLine, historyBaseLine)
	webSiteId = webSiteTableId
}

func (s BiliSpider) GetWebSiteName() models.WebSite {
	return models.WebSite{
		WebName:          webSiteName,
		WebHost:          "https://www.bilibili.com/",
		WebAuthorBaseUrl: "https://space.bilibili.com/",
		WebVideoBaseUrl:  "https://www.bilibili.com/",
	}
}

func (s BiliSpider) GetVideoList(result chan<- models.Video, closeChan chan<- models.TaskClose) {
	userEndBaseLine := make([]models.UserBaseLine, 0)
	for userName, userCookies := range cookies.GetWebSiteUser(webSiteName) {
		if userCookies == nil {
			log.ErrorLog.Printf("%s用户未初始化cookies", userName)
			continue
		}
		userEndBaseLine = append(userEndBaseLine, getUserFollowAuthorVideo(result, userCookies))
	}
	closeChan <- models.TaskClose{
		WebSite: webSiteName,
		Code:    0,
		Data:    userEndBaseLine,
	}
}
func getUserFollowAuthorVideo(result chan<- models.Video, userCookies *cookies.UserCookie) models.UserBaseLine {
	var (
		updateNumber                      int
		baseLine                          string
		ok                                bool
		dynamicBaseLine, stopVideoPutTime int64
	)
	dynamicVideoObject := dynamicVideo{userCookie: userCookies}
	dynamicBaseLine, ok = dynamicBaseLineMap[userCookies.GetUserName()]
	if !ok {
		dynamicBaseLine = 0
		dynamicBaseLineMap[userCookies.GetUserName()] = 0
	}
	if dynamicBaseLine == 0 {
		updateNumber = defaultUpdateNumber
	}
	//log.Info.Printf("%s用户开始获取关注用户视频%d", userCookies.GetUserName(), dynamicBaseLine)
	var pushTime time.Time
	breakFlag := true
	baseLine = ""
	for breakFlag {
		response := dynamicVideoObject.getResponse(0, 0, baseLine)
		if response == nil {
			log.ErrorLog.Println("请求结果错误")
			breakFlag = false
			continue
		}
		for _, info := range response.Data.Items {
			switch info.IdStr.(type) {
			case string:
				baseLine = info.IdStr.(string)
			case int:
				a := info.IdStr.(int)
				baseLine = strconv.Itoa(a)
			default:
				log.ErrorLog.Print("未知的Baseline: ", info.IdStr)
				log.ErrorLog.Println("更新基线：", baseLine)
				continue
			}

			if dynamicBaseLine >= info.Modules.ModuleAuthor.PubTs {
				breakFlag = false
				stopVideoPutTime = info.Modules.ModuleAuthor.PubTs
				break
			}

			pushTime = time.Unix(info.Modules.ModuleAuthor.PubTs, 0)
			result <- models.Video{
				WebSiteId:  webSiteId,
				Title:      info.Modules.ModuleDynamic.Major.Archive.Title,
				VideoDesc:  info.Modules.ModuleDynamic.Major.Archive.Desc,
				Duration:   HourAndMinutesAndSecondsToSeconds(info.Modules.ModuleDynamic.Major.Archive.DurationText),
				Uuid:       info.Modules.ModuleDynamic.Major.Archive.Bvid,
				Url:        info.Modules.ModuleDynamic.Major.Archive.JumpUrl,
				CoverUrl:   info.Modules.ModuleDynamic.Major.Archive.Cover,
				UploadTime: &pushTime,
				Authors: []models.VideoAuthor{
					{Contribute: "UP主", AuthorUUID: strconv.Itoa(info.Modules.ModuleAuthor.Mid)},
				},
				StructAuthor: []models.Author{
					{
						WebSiteId:    webSiteId,
						AuthorName:   info.Modules.ModuleAuthor.Name,
						AuthorWebUid: strconv.Itoa(info.Modules.ModuleAuthor.Mid),
						Avatar:       info.Modules.ModuleAuthor.Face,
					},
				},
			}

			if dynamicBaseLine == 0 {
				updateNumber--
				if updateNumber == 0 {
					breakFlag = false
					stopVideoPutTime = info.Modules.ModuleAuthor.PubTs
					break
				}
			}
		}
		if breakFlag {
			time.Sleep(time.Second * 5)
		}
	}
	dynamicBaseLineMap[userCookies.GetUserName()] = stopVideoPutTime
	return models.UserBaseLine{AuthorId: userCookies.GetDBPrimaryKeyId(), EndBaseLine: strconv.FormatInt(stopVideoPutTime, 10)}
}

func (s BiliSpider) GetAuthorDynamic(author int, baseOffset string) map[string]string {
	result := make(map[string]string)
	offset := baseOffset
	var (
		ok      bool
		_offset string
	)
	defaultUser := cookies.NewDefaultUserCookie(webSiteName)
	dynamicVideoObject := dynamicVideo{
		userCookie: defaultUser,
	}
	for {
		response := dynamicVideoObject.getResponse(0, author, offset)
		if response == nil {
			break
		}
		da, _ := json.Marshal(response)
		result[response.Data.Offset] = string(da)

		for _, info := range response.Data.Items {
			_offset, ok = info.IdStr.(string)
			if !ok {
				a, ok := info.IdStr.(int)
				if ok {
					_offset = strconv.Itoa(a)
				} else {
					break
				}
			}
			if baseOffset == _offset {
				break
			}
		}
	}

	return result
}

func (s BiliSpider) GetAuthorVideoList(author string, startPageIndex, endPageIndex int) map[int]VideoListPageResponse {
	result := make(map[int]VideoListPageResponse)
	defaultUser := cookies.NewDefaultUserCookie(webSiteName)
	video := videoListPage{
		userCookie: defaultUser,
	}
	for {
		response := video.getResponse(author, startPageIndex)
		if response == nil {
			break
		}
		result[startPageIndex] = *response
		startPageIndex++
		if startPageIndex == endPageIndex {
			break
		}
	}
	return result

}

func (s BiliSpider) GetVideoHistoryList(VideoHistoryChan chan<- models.Video, VideoHistoryCloseChan chan<- models.TaskClose) {
	userEndBaseLine := make([]models.UserBaseLine, 0)
	if historyRequestSave == nil {
		historyRequestSave = make(map[string]utils.WriteFile)
	}
	for userName, userCookies := range cookies.GetWebSiteUser(webSiteName) {
		if userCookies == nil {
			log.ErrorLog.Printf("%s用户未初始化cookies", userName)
			continue
		}
		_, ok := historyRequestSave[userName]
		if !ok {
			historyRequestSave[userName] = utils.WriteFile{
				FolderPrefix:   []string{baseStruct.RootPath, "bilbilHistoryFile"},
				FileNamePrefix: userName,
			}
		}
		userEndBaseLine = append(userEndBaseLine, getUserViewVideoHistory(VideoHistoryChan, userCookies))
	}
	VideoHistoryCloseChan <- models.TaskClose{
		WebSite: webSiteName,
		Code:    0,
		Data:    userEndBaseLine,
	}
}
func getUserViewVideoHistory(VideoHistoryChan chan<- models.Video, userCookies *cookies.UserCookie) models.UserBaseLine {
	var (
		maxNumber                                     = 100
		newestTimestamp, lastHistoryTimestamp, viewAt int64
		spiderAccount, ok                             bool
	)
	history := historyRequest{userCookie: userCookies}
	lastHistoryTimestamp, ok = historyBaseLineMap[userCookies.GetUserName()]
	if !ok {
		lastHistoryTimestamp = 0
		historyBaseLineMap[userCookies.GetUserName()] = 0
	}
	println("lastHistoryTimestamp: ", lastHistoryTimestamp)
	spiderAccount = true
	for spiderAccount {
		data := history.getResponse(0, viewAt, "archive")
		if data == nil {
			log.Info.Printf("b站%s账号爬取历史记录请求异常退出: ", userCookies.GetUserName())
			spiderAccount = false
			continue
		}
		saveFile := historyRequestSave[userCookies.GetUserName()]
		bData, _ := json.Marshal(data)
		saveFile.WriteLine(bData)
		if viewAt == 0 {
			newestTimestamp = data.Data.List[0].ViewAt
		}
		if data.Data.Cursor.Max == 0 || data.Data.Cursor.ViewAt == 0 {
			log.Info.Printf("b站%s账号爬取历史完成，爬取到%d时间", userCookies.GetUserName(), newestTimestamp)
			spiderAccount = false
			continue
		}
		viewAt = data.Data.Cursor.ViewAt
		var pushTime time.Time
		for _, info := range data.Data.List {
			if info.ViewAt <= lastHistoryTimestamp || maxNumber == 0 {
				log.Info.Printf("b站%s账号爬取历史完成，爬取到%d时间", userCookies.GetUserName(), newestTimestamp)
				spiderAccount = false
				break
			}
			switch info.Badge {
			// 稿件视频 / 剧集 / 笔记 / 纪录片 / 专栏 / 国创 / 番剧
			case "": // 稿件视频
				pushTime = time.Unix(info.ViewAt, 0)
				VideoHistoryChan <- models.Video{
					WebSiteId: webSiteId,
					Title:     info.Title,
					Uuid:      info.History.Bvid,
					CoverUrl:  info.Cover,
					Authors: []models.VideoAuthor{
						{Contribute: "UP主", AuthorUUID: strconv.FormatInt(info.AuthorMid, 10), Uuid: info.History.Bvid},
					},
					StructAuthor: []models.Author{
						{
							AuthorWebUid: strconv.FormatInt(info.AuthorMid, 10),
							AuthorName:   info.AuthorName,
							WebSiteId:    webSiteId,
							Avatar:       info.AuthorFace,
						},
					},
					ViewHistory: []models.VideoHistory{
						{ViewTime: pushTime, WebSiteId: webSiteId, WebUUID: info.History.Bvid, Duration: info.Duration, AuthorId: userCookies.GetDBPrimaryKeyId()},
					},
				}
			case "剧集":
			case "笔记":
			case "纪录片":
			case "专栏":
			case "国创":
			case "番剧":
			case "综艺":
			case "live":
				log.Info.Printf("未处理的历史记录 %v\n", info)
				continue
			default:
				log.Info.Printf("未知类型的历史记录 %v\n", info)
				continue
			}

			if lastHistoryTimestamp == 0 {
				maxNumber--
			}

		}
		if spiderAccount {
			time.Sleep(time.Second)
		}
	}
	if newestTimestamp > 0 {
		historyBaseLineMap[userCookies.GetUserName()] = newestTimestamp
	}
	return models.UserBaseLine{AuthorId: userCookies.GetDBPrimaryKeyId(), EndBaseLine: strconv.FormatInt(newestTimestamp, 10)}
}

func SaveVideoHistoryList() {
	var (
		maxNumber            = 100
		newestTimestamp      int64
		lastHistoryTimestamp int64
		business             string
		viewAt               int64
	)
	history := historyRequest{}

	a, err := os.ReadFile(path.Join(baseStruct.RootPath, "bilbilHistoryFile", "bilbilHistoryBaseLine"))
	if err != nil {
		println("读取历史记录基线错误")
		return
	}
	//defer os.WriteFile(path.Join(baseStruct.RootPath, "bilbilHistoryFile", "bilbilHistoryBaseLine"), []byte(strconv.FormatInt(newestTimestamp, 10)), 0644)
	lastHistoryTimestamp, _ = strconv.ParseInt(string(a), 10, 64)
	log.Info.Println("lastHistoryTimestamp: ", lastHistoryTimestamp)
	business = ""
	fileIndex := 1
	file := utils.WriteFile{
		FolderPrefix: []string{baseStruct.RootPath, "bilbilHistoryFile"},
		FileName: func(lastFileName string) string {
			if lastFileName == "" {
				return "bilibiliHistory"
			}
			fileIndex++
			return fmt.Sprintf("bilibiliHistory-%d", fileIndex)
		},
	}
	defer file.Close()
	for {
		data := history.getResponse(0, viewAt, business)
		if data == nil {
			log.Info.Println("退出: ", newestTimestamp)
			os.WriteFile(path.Join(baseStruct.RootPath, "bilbilHistoryFile", "bilbilHistoryBaseLine"), []byte(strconv.FormatInt(newestTimestamp, 10)), 0644)
			return
		}
		bData, _ := json.Marshal(data)
		file.WriteLine(bData)
		if newestTimestamp == 0 {
			newestTimestamp = data.Data.List[0].ViewAt
		}
		if data.Data.Cursor.Max == 0 || data.Data.Cursor.ViewAt == 0 {
			os.WriteFile(path.Join(baseStruct.RootPath, "bilbilHistoryFile", "bilbilHistoryBaseLine"), []byte(strconv.FormatInt(newestTimestamp, 10)), 0644)
			log.Info.Println("退出: ", newestTimestamp)
			return
		}

		viewAt = data.Data.Cursor.ViewAt
		business = data.Data.Cursor.Business
		for _, info := range data.Data.List {
			if info.ViewAt < lastHistoryTimestamp || maxNumber == 0 {
				os.WriteFile(path.Join(baseStruct.RootPath, "bilbilHistoryFile", "bilbilHistoryBaseLine"), []byte(strconv.FormatInt(newestTimestamp, 10)), 0644)
				log.Info.Println("退出: ", newestTimestamp)
				return
			}

			if lastHistoryTimestamp == 0 {
				maxNumber--
			}

		}
		time.Sleep(time.Second)
	}
}

type WaitUpdateCollect struct {
	BvId          int64
	CollectId     int64
	CollectNumber int64
}

type NewCollect struct {
	Collect []WaitUpdateCollect
	Season  []WaitUpdateCollect
}

// GetCollectList 获取收藏夹和专栏列表,不包含里面的视频信息
func (s BiliSpider) GetCollectList() NewCollect {
	result := new(NewCollect)
	var (
		collectInfo models.Collect
		err         error
		PageIndex   int
	)
	PageIndex = 1
	for _, userCookies := range cookies.GetWebSiteUser(webSiteName) {
		mid := userCookies.GetCookiesKeyValue("DedeUserID")
		for {
			a := getCollectList(mid, PageIndex, userCookies)
			if a != nil {
				for _, info := range a.Data.List {
					collectInfo = models.Collect{}
					err = models.GormDB.Where("`type`=? and bv_id=?", 1, info.Id).Find(&collectInfo).Error
					if err != nil {
						log.ErrorLog.Printf("GetCollectList查询Collect表出错")
						continue
					}
					if collectInfo.Id == 0 {
						collectInfo.Type = 1
						collectInfo.BvId = info.Id
						collectInfo.Name = info.Title
						err = models.GormDB.Create(&collectInfo).Error
					}
					if err == nil {
						result.Collect = append(result.Collect, WaitUpdateCollect{BvId: info.Id, CollectId: collectInfo.Id, CollectNumber: info.MediaCount})
					}
				}
			} else {
				break
			}
			if a.Data.Count <= (PageIndex * 50) {
				break
			}
			PageIndex++
		}
		time.Sleep(time.Second)
		PageIndex = 1
		for {
			b := subscriptionList(mid, PageIndex, userCookies)
			if b != nil {
				for _, info := range b.Data.List {
					collectInfo = models.Collect{}
					err = models.GormDB.Where("`type`=? and bv_id=?", 2, info.Id).Find(&collectInfo).Error
					if err != nil {
						log.ErrorLog.Printf("GetCollectList查询Collect表出错")
						continue
					}
					if info.Mid == 0 {
						if collectInfo.Id > 0 && !strings.Contains(collectInfo.Name, "合集已失效") {
							// 该合集以失效，更新数据
							collectInfo.Name += "(合集已失效)"
							models.GormDB.Model(&collectInfo).Update("name", collectInfo.Name)
						}
						continue
					}
					if collectInfo.Id == 0 {
						collectInfo.Type = 2
						collectInfo.BvId = info.Id
						collectInfo.Name = info.Title
						err = models.GormDB.Create(&collectInfo).Error
					}
					if err == nil {
						result.Season = append(result.Season, WaitUpdateCollect{BvId: info.Id, CollectId: collectInfo.Id, CollectNumber: info.MediaCount})
					}
				}
			} else {
				break
			}
			if b.Data.Count <= (PageIndex * 50) {
				break
			}
			PageIndex++
		}
	}

	return *result
}

// GetCollectAllVideo 获取收藏夹里面的视频信息
func (s BiliSpider) GetCollectAllVideo(collectId int64, maxPage int) []CollectVideoDetailInfo {
	var (
		response    *collectVideoDetailResponse
		videoNumber int
		page        = 1
		result      []CollectVideoDetailInfo
	)
	for _, userCookies := range cookies.GetWebSiteUser(webSiteName) {
		for {
			response = getCollectVideoInfo(collectId, page, userCookies)
			if videoNumber == 0 {
				videoNumber = response.Data.Info.MediaCount
				if maxPage == 0 {
					if videoNumber%20 == 0 {
						maxPage = videoNumber / 20
					} else {
						maxPage = (videoNumber / 20) + 1
					}
				}
			}
			for _, info := range response.Data.Medias {
				result = append(result, info)
			}
			if page >= maxPage {
				break
			}
			page++
			time.Sleep(time.Second)
		}
	}
	return result
}

// GetSeasonAllVideo 获取专栏里面的视频信息
func (s BiliSpider) GetSeasonAllVideo(seasonId int64) []SeasonAllVideoDetailInfo {
	var (
		response *seasonAllVideoDetailResponse
		result   []SeasonAllVideoDetailInfo
	)
	response = getSeasonVideoInfo(seasonId, 1)
	if response == nil {
		println(seasonId, "合集返回空")
		return result
	}
	for _, info := range response.Data.Medias {
		result = append(result, info)
	}

	return result

}

func (s BiliSpider) GetFollowingList(resultChan chan<- baseStruct.FollowInfo, closeChan chan<- int64) {
	var (
		total   = 0
		maxPage = 1
		f       followings
		userId  int64
	)
	f = followings{
		pageNumber: 1,
	}
	for fileName, userCookies := range cookies.GetWebSiteUser(webSiteName) {
		userId = userCookies.GetDBPrimaryKeyId()
		if userId == 0 {
			log.ErrorLog.Printf("%s用户不存在,请手动添加", fileName)
			continue
		}
		f.pageNumber = 1
		f.userCookies = userCookies
		maxPage = 1
		total = 0
		log.Info.Printf("%s用户的id是:%d", fileName, userId)
		for {
			response := f.getResponse(0)
			if response == nil {
				response = &followingsResponse{}
			}
			if total == 0 {
				total = response.Data.Total
				if total%20 == 0 {
					maxPage = total / 20
				} else {
					maxPage = (total / 20) + 1
				}
			}
			for _, info := range response.Data.List {
				followTime := time.Unix(info.Mtime, 0)
				resultChan <- baseStruct.FollowInfo{
					WebSiteId:  webSiteId,
					UserId:     userId,
					AuthorName: info.Uname,
					AuthorUUID: strconv.FormatInt(info.Mid, 10),
					Avatar:     info.Face,
					AuthorDesc: info.Sign,
					FollowTime: &followTime,
				}
			}
			if f.pageNumber >= maxPage {
				break
			}
			f.pageNumber++
			time.Sleep(time.Second * 1)
		}
	}

	closeChan <- 0
	return
}

func (s BiliSpider) GetSelfInfo(cookiesContext string) models.AccountInfo {
	userCookies := cookies.NewTemporaryUserCookie(webSiteName, cookiesContext)
	data, err := getSelfInfo(userCookies.GetCookiesKeyValue(selfUserMid), cookiesContext)
	if err != nil {
		return nil
	}
	return data
}
