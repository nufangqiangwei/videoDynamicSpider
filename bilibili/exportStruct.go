package bilibili

import (
	"encoding/json"
	"strconv"
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

func (s BiliSpider) GetVideoList(latestBaseline string, result chan<- baseStruct.VideoInfo, closeChan chan<- baseStruct.TaskClose) {
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

			result <- baseStruct.VideoInfo{
				WebSite:    "bilibili",
				Title:      info.Modules.ModuleDynamic.Major.Archive.Title,
				Desc:       info.Modules.ModuleDynamic.Major.Archive.Desc,
				Duration:   HourAndMinutesAndSecondsToSeconds(info.Modules.ModuleDynamic.Major.Archive.DurationText),
				VideoUuid:  info.Modules.ModuleDynamic.Major.Archive.Bvid,
				Url:        info.Modules.ModuleDynamic.Major.Archive.JumpUrl,
				CoverUrl:   info.Modules.ModuleDynamic.Major.Archive.Cover,
				AuthorUuid: strconv.Itoa(info.Modules.ModuleAuthor.Mid),
				AuthorName: info.Modules.ModuleAuthor.Name,
				AuthorUrl:  info.Modules.ModuleAuthor.JumpUrl,
				PushTime:   time.Unix(info.Modules.ModuleAuthor.PubTs, 0),
				Baseline:   Baseline,
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

func (s BiliSpider) GetVideoHistoryList(lastHistoryTimestamp int64, VideoHistoryChan chan<- baseStruct.VideoInfo, VideoHistoryCloseChan chan<- int64) {
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
				VideoHistoryChan <- baseStruct.VideoInfo{
					WebSite:    "bilibili",
					Title:      info.Title,
					VideoUuid:  info.History.Bvid,
					AuthorUuid: strconv.FormatInt(info.AuthorMid, 10),
					AuthorName: info.AuthorName,
					PushTime:   time.Unix(info.ViewAt, 0),
				}
			case "剧集":
			case "笔记":
			case "纪录片":
			case "专栏":
			case "国创":
			case "番剧":
			case "综艺":
			case "live":
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

type NewCollect struct {
	Collect []int64
	Season  []int64
}

// GetCollectList 获取收藏夹和专栏列表,不包含里面的视频信息
func (s BiliSpider) GetCollectList() NewCollect {
	a := getCollectList("10932398")
	result := new(NewCollect)
	var x models.Collect
	if a != nil {
		for _, info := range a.Data.List {
			x = models.Collect{
				Type: 1,
				BvId: info.Id,
				Name: info.Title,
			}
			if x.Save() {
				result.Collect = append(result.Collect, x.BvId)
			}
		}
	}
	time.Sleep(time.Second)
	b := subscriptionList("10932398")
	if b != nil {
		for _, info := range b.Data.List {
			x = models.Collect{
				Type: 2,
				BvId: info.Id,
				Name: info.Title,
			}
			if x.Save() {
				result.Season = append(result.Season, x.BvId)
			}
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
