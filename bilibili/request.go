package bilibili

import (
	"time"
	"videoDynamicAcquisition/baseStruct"
)

var (
	Spider             BiliSpider
	bilibiliCookies    cookies
	dynamicVideoObject dynamicVideo
	dynamicBaseUrl     = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed"
	followingsBseUrl   = "https://api.bilibili.com/x/relation/followings"
	latestBaseline     = "" // 836201790065082504
	wbiSignObj         = wbiSign{}
)

func init() {
	bilibiliCookies = cookies{}
	Spider = BiliSpider{
		VideoHistoryChan:      make(chan baseStruct.VideoInfo, 100),
		VideoHistoryCloseChan: make(chan string, 10),
	}
	dynamicVideoObject = dynamicVideo{}
	bilibiliCookies.readFile()
	wbiSignObj.lastUpdateTime = time.Now()
}
