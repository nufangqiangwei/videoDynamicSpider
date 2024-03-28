package bilibili

import (
	"strconv"
	"strings"
	"videoDynamicAcquisition/cookies"
)

// HourAndMinutesAndSecondsToSeconds 时长字符串转秒数
// 时长字符串例子： [00:19, 02:45, 01:30:14]
func HourAndMinutesAndSecondsToSeconds(time string) int {
	split := strings.Split(time, ":")
	if len(split) == 2 {
		minutes, _ := strconv.Atoi(split[0])
		seconds, _ := strconv.Atoi(split[1])
		return minutes*60 + seconds
	} else if len(split) == 3 {
		hour, _ := strconv.Atoi(split[0])
		minutes, _ := strconv.Atoi(split[1])
		seconds, _ := strconv.Atoi(split[2])
		return hour*3600 + minutes*60 + seconds
	}
	return 0
}

const defaultUserCookies = "buvid3=C291FA4B-1420-D9DD-EDF2-59BD54E9FF2484319infoc; b_nut=1687748884; CURRENT_FNVAL=4048; _uuid=11054346D-34210-64EC-65F1-D743C7733410F85132infoc; buvid_fp=C291FA4B-1420-D9DD-EDF2-59BD54E9FF2484319infoc; buvid4=D2B0F4FD-B1F6-D046-9097-260FFC188E3085245-023062611-5%2FGnnPQYM6GlbwD8D1EmcA%3D%3D; rpdid=|(J|~)~u)kl)0J'uY)~~R)))l; header_theme_version=CLOSE; home_feed_column=5; browser_resolution=1920-967; hit-new-style-dyn=1; hit-dyn-v2=1; PVID=1; LIVE_BUVID=AUTO3316938052076221; enable_web_push=DISABLE; fingerprint=aea90f7d07f026587b224ad4c246f7c5; buvid_fp_plain=undefined; CURRENT_QUALITY=80; bp_video_offset_3546632784185451=913508315169816585; sid=q2v02b6w; FEED_LIVE_VERSION=V8; share_source_origin=WEIXIN; b_lsid=431012B69_18E84726511; bsource=share_source_weixinchat; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTE4NzgyODAsImlhdCI6MTcxMTYxOTAyMCwicGx0IjotMX0.cJ0b8MuU6dKzwX_L9QDL1Rt3DY8DJmvty-RRsSPtODg; bili_ticket_expires=1711878220"

func getUser() *cookies.UserCookie {
	for _, user := range cookies.GetWebSiteUser(webSiteName) {
		return user
	}
	return getDefaultUser()
}
func getDefaultUser() *cookies.UserCookie {
	return cookies.NewTemporaryUserCookie(webSiteName, defaultUserCookies)
}
