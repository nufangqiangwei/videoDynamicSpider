package bilibili

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/proxy"
)

const (
	rankingUrl     = "https://api.bilibili.com/x/web-interface/popular"
	weekRankingUrl = "https://api.bilibili.com/x/web-interface/popular/series/one"
)

var weekRankingStartTime = time.Date(2019, time.March, 22, 0, 0, 0, 0, time.UTC)

// 热门视频获取
// api文档 https://socialsisteryi.github.io/bilibili-API-collect/docs/video_ranking/popular.html
// api https://api.bilibili.com/x/web-interface/popular
type RankIngResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		List   []RankIngResponseVideoInfo `json:"list"`
		NoMore bool                       `json:"no_more"`
	} `json:"data"`
}
type RankIngResponseVideoInfo struct {
	Aid       int64  `json:"aid"`
	Videos    int    `json:"videos"`
	Tid       int64  `json:"tid"`
	Tname     string `json:"tname"`
	Copyright int    `json:"copyright"`
	Pic       string `json:"pic"`
	Title     string `json:"title"`
	Pubdate   int64  `json:"pubdate"`
	Ctime     int64  `json:"ctime"`
	Desc      string `json:"desc"`
	State     int    `json:"state"`
	Duration  int    `json:"duration"`
	Rights    struct {
		Bp            int `json:"bp"`
		Elec          int `json:"elec"`
		Download      int `json:"download"`
		Movie         int `json:"movie"`
		Pay           int `json:"pay"`
		Hd5           int `json:"hd5"`
		NoReprint     int `json:"no_reprint"`
		Autoplay      int `json:"autoplay"`
		UgcPay        int `json:"ugc_pay"`
		IsCooperation int `json:"is_cooperation"`
		UgcPayPreview int `json:"ugc_pay_preview"`
		NoBackground  int `json:"no_background"`
		ArcPay        int `json:"arc_pay"`
		PayFreeWatch  int `json:"pay_free_watch"`
	} `json:"rights"`
	Owner struct {
		Mid  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"owner"`
	Stat struct {
		Aid      int64 `json:"aid"`
		View     int64 `json:"view"`
		Danmaku  int64 `json:"danmaku"`
		Reply    int64 `json:"reply"`
		Favorite int64 `json:"favorite"`
		Coin     int64 `json:"coin"`
		Share    int64 `json:"share"`
		NowRank  int64 `json:"now_rank"`
		HisRank  int64 `json:"his_rank"`
		Like     int64 `json:"like"`
		Dislike  int64 `json:"dislike"`
		Vt       int64 `json:"vt"`
		Vv       int64 `json:"vv"`
	} `json:"stat"`
	Dynamic   string `json:"dynamic"`
	Cid       int    `json:"cid"`
	Dimension struct {
		Width  int `json:"width"`
		Height int `json:"height"`
		Rotate int `json:"rotate"`
	} `json:"dimension"`
	ShortLinkV2 string      `json:"short_link_v2"`
	UpFromV2    int         `json:"up_from_v2,omitempty"`
	FirstFrame  string      `json:"first_frame"`
	PubLocation string      `json:"pub_location"`
	Cover43     string      `json:"cover43"`
	Bvid        string      `json:"bvid"`
	SeasonType  int         `json:"season_type"`
	IsOgv       bool        `json:"is_ogv"`
	OgvInfo     interface{} `json:"ogv_info"`
	EnableVt    int         `json:"enable_vt"`
	AiRcmd      interface{} `json:"ai_rcmd"`
	RcmdReason  struct {
		Content    string `json:"content"`
		CornerMark int    `json:"corner_mark"`
	} `json:"rcmd_reason"`
	MissionId int `json:"mission_id,omitempty"`
	SeasonId  int `json:"season_id,omitempty"`
}

// RankIngResponse 在指针上实现responseCheck接口
func (r *RankIngResponse) getCode() int {
	return r.Code
}
func (r *RankIngResponse) bindJSON(body []byte) error {
	return json.Unmarshal(body, r)
}

/*
1 2019第1期 03.22 - 03.28
2 2019第2期 03.29 - 04.04
3 2019第3期 04.05 - 04.11
4 2019第4期 04.12 - 04.18
*/
// 计算从2019-03-22这一天开始，距离现在时间是第多少周
func getWeekRankingNumber() int {
	currentDate := time.Now().UTC()
	return int(currentDate.Sub(weekRankingStartTime).Hours() / 24 / 7)

}

// RankIng 获取指定排行榜
type RankIng struct {
	hotType string
}

func (r RankIng) getDayHotRequest(user *cookies.UserCookie, page, size int) *http.Request {
	request, _ := http.NewRequest("GET", rankingUrl, nil)
	if user != nil {
		request.Header.Add("Cookie", user.GetCookies())
	}
	request.Header.Add("User-Agent", userAgent)
	q := request.URL.Query()
	q.Add("ps", strconv.Itoa(size))
	q.Add("pn", strconv.Itoa(page))
	request.URL.RawQuery = q.Encode()
	return request
}
func (r RankIng) getWeekHotVideoList(user *cookies.UserCookie) *http.Request {
	request, _ := http.NewRequest("GET", weekRankingUrl, nil)
	request.Header.Add("User-Agent", userAgent)
	if user != nil {
		request.Header.Add("Cookie", user.GetCookies())
	}
	q := request.URL.Query()
	q.Add("number", strconv.Itoa(getWeekRankingNumber()))
	request.URL.RawQuery = q.Encode()
	return request
}
func (r RankIng) getResponse(retriesNumber int, user *cookies.UserCookie, page, size int, useProxy bool) *RankIngResponse {
	if retriesNumber > 2 {
		log.ErrorLog.Println("重试次数过多")
		return nil
	}
	var request *http.Request
	if r.hotType == "day" {
		request = r.getDayHotRequest(user, page, size)
	} else if r.hotType == "week" {
		request = r.getWeekHotVideoList(user)
	} else {
		request = r.getDayHotRequest(user, page, size)
	}

	response, err := proxy.GetClient(useProxy).Do(request)
	if err != nil {
		log.ErrorLog.Println("请求失败")
		log.ErrorLog.Println(err.Error())
		return nil
	}
	result := new(RankIngResponse)
	if user == nil {
		user = cookies.NewDefaultUserCookie(webSiteName)
	}
	err = responseCodeCheck(response, result, user)
	if err != nil {
		return nil
	}
	return result
}
