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
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

type BiliSpider struct {
}

func (s BiliSpider) GetWebSiteName() models.WebSite {
	return models.WebSite{
		WebName:          "bilibili",
		WebHost:          "https://www.bilibili.com/",
		WebAuthorBaseUrl: "https://space.bilibili.com/",
		WebVideoBaseUrl:  "https://www.bilibili.com/",
	}
}

func (s BiliSpider) GetVideoList(latestBaseline string, result chan<- models.Video, closeChan chan<- baseStruct.TaskClose) {
	var (
		intLatestBaseline int
		videoBaseLine     int
		startBaseLine     int
		err               error
		baseLine          string
		updateNumber      int
		Baseline          string
		ok                bool
		requestNumber     int
	)
	if latestBaseline == "" {
		updateNumber = 20
	} else {
		intLatestBaseline, err = strconv.Atoi(latestBaseline)
		if err != nil {
			utils.ErrorLog.Printf("intLatestBaseline 转换出错：%d\n", intLatestBaseline)
			updateNumber = 20
		}
	}
	webSite := models.WebSite{}
	models.GormDB.Where("web_name=?", "bilibili").First(&webSite)
	var pushTime time.Time
	for {
		response := dynamicVideoObject.getResponse(0, 0, baseLine)

		if response == nil {
			closeChan <- baseStruct.TaskClose{
				WebSite: "bilibili",
				Code:    0,
			}
			return
		}

		for infoIndex, info := range response.Data.Items {
			Baseline, ok = info.IdStr.(string)
			if !ok {
				a, ok := info.IdStr.(int)
				if ok {
					Baseline = strconv.Itoa(a)
					videoBaseLine = a
				} else {
					utils.ErrorLog.Print("未知的Baseline: ", info.IdStr)
					utils.ErrorLog.Println("更新基线：", baseLine)
					continue
				}
			} else {
				videoBaseLine, err = strconv.Atoi(Baseline)
				if err != nil {
					utils.ErrorLog.Println("视频的IDStr错误")
					continue
				}
			}

			if videoBaseLine <= intLatestBaseline {
				closeChan <- baseStruct.TaskClose{
					WebSite: "bilibili",
					Code:    startBaseLine,
				}
				return
			}

			if requestNumber == 0 && infoIndex == 0 {
				startBaseLine = videoBaseLine
			}
			pushTime = time.Unix(info.Modules.ModuleAuthor.PubTs, 0)
			result <- models.Video{
				WebSiteId:  webSite.Id,
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
						WebSiteId:    webSite.Id,
						AuthorName:   info.Modules.ModuleAuthor.Name,
						AuthorWebUid: strconv.Itoa(info.Modules.ModuleAuthor.Mid),
						Avatar:       info.Modules.ModuleAuthor.Face,
					},
				},
			}

			if latestBaseline == "" {
				updateNumber--
				if updateNumber == 0 {
					closeChan <- baseStruct.TaskClose{
						WebSite: "bilibili",
						Code:    startBaseLine,
					}
					return
				}
			}

		}
		baseLine = response.Data.Offset
		requestNumber++
		time.Sleep(time.Second * 10)
	}
}

func (s BiliSpider) GetAuthorDynamic(author int, baseOffset string) map[string]string {
	result := make(map[string]string)
	offset := baseOffset
	var (
		ok      bool
		_offset string
	)
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

func (s BiliSpider) GetAuthorVideoList(author string, startPageIndex, endPageIndex int) map[string]string {
	result := make(map[string]string)
	video := videoListPage{}
	for {
		response := video.getResponse(author, startPageIndex)
		if response == nil {
			break
		}
		da, _ := json.Marshal(response)
		result[strconv.Itoa(startPageIndex)] = string(da)
		startPageIndex++
		if startPageIndex == endPageIndex {
			break
		}
	}
	return result

}

func (s BiliSpider) GetVideoHistoryList(lastHistoryTimestamp int64, VideoHistoryChan chan<- models.Video, VideoHistoryCloseChan chan<- int64) {
	history := historyRequest{}
	var (
		maxNumber       = 100
		newestTimestamp int64
		business        string
		viewAt          int64
		max             int
	)

	println("lastHistoryTimestamp: ", lastHistoryTimestamp)

	business = ""
	webSite := models.WebSite{}
	models.GormDB.Where("web_name=?", "bilibili").First(&webSite)
	var pushTime time.Time
	for {
		data := history.getResponse(max, viewAt, business)
		if data == nil {
			println("退出: ", 0)
			VideoHistoryCloseChan <- 0
			return
		}
		if viewAt == 0 {
			newestTimestamp = data.Data.List[0].ViewAt
		}
		if data.Data.Cursor.Max == 0 || data.Data.Cursor.ViewAt == 0 {
			// https://s1.hdslb.com/bfs/static/history-record/img/historyend.png
			println("退出: ", newestTimestamp)
			VideoHistoryCloseChan <- newestTimestamp
			return
		}
		max = data.Data.Cursor.Max
		viewAt = data.Data.Cursor.ViewAt
		business = data.Data.Cursor.Business
		for _, info := range data.Data.List {
			if info.ViewAt < lastHistoryTimestamp || maxNumber == 0 {
				VideoHistoryCloseChan <- newestTimestamp
				return
			}
			switch info.Badge {
			// 稿件视频 / 剧集 / 笔记 / 纪录片 / 专栏 / 国创 / 番剧
			case "": // 稿件视频
				pushTime = time.Unix(info.ViewAt, 0)
				VideoHistoryChan <- models.Video{
					WebSiteId: webSite.Id,
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
							WebSiteId:    webSite.Id,
							Avatar:       info.AuthorFace,
						},
					},
					ViewHistory: []models.VideoHistory{
						{ViewTime: pushTime, WebSiteId: webSite.Id, WebUUID: info.History.Bvid},
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
				utils.Info.Printf("未处理的历史记录 %v\n", info)
				continue
			default:
				utils.Info.Printf("未知类型的历史记录 %v\n", info)
				continue
			}

			if lastHistoryTimestamp == 0 {
				maxNumber--
			}

		}
		time.Sleep(time.Second)
	}
}

func SaveVideoHistoryList() {
	var (
		maxNumber            = 100
		newestTimestamp      int64
		lastHistoryTimestamp int64
		business             string
		viewAt               int64
		max                  int
	)
	history := historyRequest{}

	a, err := os.ReadFile(path.Join(baseStruct.RootPath, "bilbilHistoryFile", "bilbilHistoryBaseLine"))
	if err != nil {
		println("读取历史记录基线错误")
		return
	}
	defer os.WriteFile(path.Join(baseStruct.RootPath, "bilbilHistoryFile", "bilbilHistoryBaseLine"), []byte(strconv.FormatInt(newestTimestamp, 10)), 0644)
	lastHistoryTimestamp, _ = strconv.ParseInt(string(a), 10, 64)
	println("lastHistoryTimestamp: ", lastHistoryTimestamp)
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
		data := history.getResponse(max, viewAt, business)
		if data == nil {
			println("退出: ", newestTimestamp)
			return
		}
		bData, _ := json.Marshal(data)
		file.Write(bData)
		if viewAt == 0 {
			newestTimestamp = data.Data.List[0].ViewAt
		}
		if data.Data.Cursor.Max == 0 || data.Data.Cursor.ViewAt == 0 {
			// https://s1.hdslb.com/bfs/static/history-record/img/historyend.png
			println("退出: ", newestTimestamp)
			return
		}
		max = data.Data.Cursor.Max
		viewAt = data.Data.Cursor.ViewAt
		business = data.Data.Cursor.Business
		for _, info := range data.Data.List {
			if info.ViewAt < lastHistoryTimestamp || maxNumber == 0 {
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
	for {
		a := getCollectList("10932398", PageIndex)
		if a != nil {
			for _, info := range a.Data.List {
				collectInfo = models.Collect{}
				err = models.GormDB.Where("`type`=? and bv_id=?", 1, info.Id).Find(&collectInfo).Error
				if err != nil {
					utils.ErrorLog.Printf("GetCollectList查询Collect表出错")
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
		b := subscriptionList("10932398", PageIndex)
		if b != nil {
			for _, info := range b.Data.List {
				collectInfo = models.Collect{}
				err = models.GormDB.Where("`type`=? and bv_id=?", 2, info.Id).Find(&collectInfo).Error
				if err != nil {
					utils.ErrorLog.Printf("GetCollectList查询Collect表出错")
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
	for {
		response = getCollectVideoInfo(collectId, page)
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

func (s BiliSpider) GetFollowingList(resultChan chan<- FollowingUP, closeChan chan<- int64) {
	var (
		total   = 0
		maxPage = 1
		f       followings
	)
	f = followings{
		pageNumber: 1,
	}
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
			resultChan <- info
		}
		if f.pageNumber >= maxPage {
			break
		}
		f.pageNumber++
		time.Sleep(time.Second * 3)
	}
	closeChan <- 0
	return
}
