package bilibili

import "time"

var (
	Bilibili           BilibiliSpider
	bilibiliCookies    cookies
	dynamicVideoObject dynamicVideo
	dynamicBaseUrl     = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed"
	followingsBseUrl   = "https://api.bilibili.com/x/relation/followings"
	latestBaseline     = "" // 836201790065082504
	wbiSignObj         = wbiSign{}
)

func init() {
	bilibiliCookies = cookies{}
	Bilibili = BilibiliSpider{}
	dynamicVideoObject = dynamicVideo{}
	bilibiliCookies.readFile()
	wbiSignObj.lastUpdateTime = time.Now()
}
