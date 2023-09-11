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

func (s BiliSpider) GetVideoList() []baseStruct.VideoInfo {
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
			return result
		}
		if response.Data.Items == nil {
			errorRetriesNumber++
			if errorRetriesNumber > 3 {
				utils.ErrorLog.Println("多次获取数据失败，退出")
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
					utils.ErrorLog.Println("未知的Baseline: ", info.IdStr)
				}
			}
			if len(result) == 0 {
				latestBaseline = Baseline
				println("aaaaaaaaaaaaaaaaaaaaaaaaaa", latestBaseline)
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
