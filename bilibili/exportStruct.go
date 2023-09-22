package bilibili

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

type BiliSpider struct {
	VideoHistoryChan      chan baseStruct.VideoInfo
	VideoHistoryCloseChan chan int64
}

func (s BiliSpider) GetWebSiteName() models.WebSite {
	return models.WebSite{
		WebName:          "bilibili",
		WebHost:          "https://www.bilibili.com/",
		WebAuthorBaseUrl: "https://space.bilibili.com/",
		WebVideoBaseUrl:  "https://www.bilibili.com/",
	}
}

func (s BiliSpider) GetVideoList(latestBaseline string) []baseStruct.VideoInfo {
	var updateNumber int
	if latestBaseline == "" {
		updateNumber = 20
	} else {
		updateNumber = dynamicVideoObject.getUpdateVideoNumber(latestBaseline)
	}

	utils.Info.Printf("updateNumber: %d\n", updateNumber)
	pageNumber := 1
	result := make([]baseStruct.VideoInfo, 0)
	if updateNumber == 0 {
		return result
	}
	baseLine := latestBaseline
	errorRetriesNumber := 0
	for updateNumber >= 0 {
		response := dynamicVideoObject.getResponse(0, 0, baseLine)

		if response == nil {
			result = []baseStruct.VideoInfo{}
			return result
		}
		if response.Data.Items == nil {
			errorRetriesNumber++
			if errorRetriesNumber > 3 {
				utils.ErrorLog.Println("多次获取数据失败，退出")
				result = []baseStruct.VideoInfo{}
				return result
			}
			continue
		}

		for _, info := range response.Data.Items {
			var Baseline string
			Baseline, ok := info.IdStr.(string)
			if !ok {
				a, ok := info.IdStr.(int)
				if ok {
					Baseline = strconv.Itoa(a)
				} else {
					utils.ErrorLog.Print("未知的Baseline: ", info.IdStr)
					utils.ErrorLog.Println("更新基线：", baseLine)
					continue
				}
			}
			result = append(result, baseStruct.VideoInfo{
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
			})
			updateNumber--
			if updateNumber == 0 {
				break
			}
		}
		pageNumber++
		baseLine = response.Data.Offset
		time.Sleep(time.Second * 10)
	}

	return result
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

func (s BiliSpider) GetVideoHistoryList(lastHistoryTimestamp int64) {
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
			s.VideoHistoryCloseChan <- 0
			return
		}
		if viewAt == 0 {
			newestTimestamp = data.Data.List[0].ViewAt
		}
		if data.Data.Cursor.Max == 0 || data.Data.Cursor.ViewAt == 0 {
			// https://s1.hdslb.com/bfs/static/history-record/img/historyend.png
			println("退出: ", newestTimestamp)
			s.VideoHistoryCloseChan <- newestTimestamp
			return
		}
		max = data.Data.Cursor.Max
		viewAt = data.Data.Cursor.ViewAt
		business = data.Data.Cursor.Business
		for _, info := range data.Data.List {
			if info.ViewAt < lastHistoryTimestamp || maxNumber == 0 {
				println("退出: ", viewAt)
				s.VideoHistoryCloseChan <- newestTimestamp
				return
			}
			s.VideoHistoryChan <- baseStruct.VideoInfo{
				WebSite:    "bilibili",
				Title:      info.Title,
				VideoUuid:  info.History.Bvid,
				AuthorUuid: strconv.FormatInt(info.AuthorMid, 10),
				AuthorName: info.AuthorName,
				PushTime:   time.Unix(info.ViewAt, 0),
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

// GetCollectList 获取收藏夹和专栏列表
func (s BiliSpider) GetCollectList(db *sql.DB) NewCollect {
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
			if x.CreateOrQuery(db) {
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
			if x.CreateOrQuery(db) {
				result.Season = append(result.Season, x.BvId)
			}
		}
	}
	return *result
}

func (s BiliSpider) GetCollectAllVideo(collectId int64) []CollectVideoDetailInfo {
	var (
		response             *collectVideoDetailResponse
		videoNumber, maxPage int
		page                 = 1
		result               []CollectVideoDetailInfo
	)
	for {
		response = getCollectVideoInfo(collectId, page)
		if videoNumber == 0 {
			videoNumber = response.Data.Info.MediaCount
			if videoNumber%20 == 0 {
				maxPage = videoNumber / 20
			} else {
				maxPage = (videoNumber / 20) + 1
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
