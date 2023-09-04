package bilibili

import (
	"fmt"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
)

type BilibiliSpider struct {
	lastFlushCookiesTime time.Time
	cookiesFail          bool
}

func MakeBilibiliSpider() BilibiliSpider {
	bilibiliCookies = cookies{}
	bilibili := BilibiliSpider{}
	dynamicVideoObject = dynamicVideo{}
	bilibiliCookies.readFile()
	return bilibili
}

func (bilibili BilibiliSpider) GetWebSiteName() models.WebSite {
	return models.WebSite{
		WebName:          "bilibili",
		WebHost:          "https://www.bilibili.com/",
		WebAuthorBaseUrl: "https://space.bilibili.com/",
		WebVideoBaseUrl:  "https://www.bilibili.com/",
	}
}

func (bilibili BilibiliSpider) GetVideoList() []baseStruct.VideoInfo {
	var updateNumber int
	if latestBaseline == "" {
		updateNumber = 20
	} else {
		updateNumber = dynamicVideoObject.getUpdateVideoNumber(latestBaseline)
	}
	fmt.Printf("updateNumber: %d\n", updateNumber)
	// 一页返回最多十四条数据，需要计算最大页数

	pageNumber := 1
	result := make([]baseStruct.VideoInfo, 0)
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
				println("多次获取数据失败，退出")

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

	if len(result) == updateNumber {
		latestBaseline = result[0].Baseline
	}

	return result
}
