package bilibili

import (
	"encoding/json"
	"net/http"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/proxy"
)

type WaitWatchVideoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Count int              `json:"count"`
		List  []WaitWatchVideo `json:"list"`
	} `json:"data"`
}

type WaitWatchVideo struct {
	Aid       int    `json:"aid"`
	Videos    int    `json:"videos"`
	Tid       int    `json:"tid"`
	Tname     string `json:"tname"`
	Copyright int    `json:"copyright"`
	Pic       string `json:"pic"`
	Title     string `json:"title"`
	Pubdate   int    `json:"pubdate"`
	Ctime     int    `json:"ctime"`
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
		Aid      int `json:"aid"`
		View     int `json:"view"`
		Danmaku  int `json:"danmaku"`
		Reply    int `json:"reply"`
		Favorite int `json:"favorite"`
		Coin     int `json:"coin"`
		Share    int `json:"share"`
		NowRank  int `json:"now_rank"`
		HisRank  int `json:"his_rank"`
		Like     int `json:"like"`
		Dislike  int `json:"dislike"`
		Vt       int `json:"vt"`
		Vv       int `json:"vv"`
	} `json:"stat"`
	Dynamic   string `json:"dynamic"`
	Dimension struct {
		Width  int `json:"width"`
		Height int `json:"height"`
		Rotate int `json:"rotate"`
	} `json:"dimension"`
	SeasonId    int    `json:"season_id,omitempty"`
	ShortLinkV2 string `json:"short_link_v2"`
	FirstFrame  string `json:"first_frame,omitempty"`
	PubLocation string `json:"pub_location,omitempty"`
	Cover43     string `json:"cover43"`
	Page        struct {
		Cid       int    `json:"cid"`
		Page      int    `json:"page"`
		From      string `json:"from"`
		Part      string `json:"part"`
		Duration  int    `json:"duration"`
		Vid       string `json:"vid"`
		Weblink   string `json:"weblink"`
		Dimension struct {
			Width  int `json:"width"`
			Height int `json:"height"`
			Rotate int `json:"rotate"`
		} `json:"dimension"`
		FirstFrame string `json:"first_frame,omitempty"`
	} `json:"page"`
	Count         int    `json:"count"`
	Cid           int    `json:"cid"`
	Progress      int    `json:"progress"`
	AddAt         int    `json:"add_at"`
	Bvid          string `json:"bvid"`
	Uri           string `json:"uri"`
	EnableVt      int    `json:"enable_vt"`
	ViewText1     string `json:"view_text_1"`
	CardType      int    `json:"card_type"`
	LeftIconType  int    `json:"left_icon_type"`
	LeftText      string `json:"left_text"`
	RightIconType int    `json:"right_icon_type"`
	RightText     string `json:"right_text"`
	ArcState      int    `json:"arc_state"`
	PgcLabel      string `json:"pgc_label"`
	ShowUp        bool   `json:"show_up"`
	ForbidFav     bool   `json:"forbid_fav"`
	ForbidSort    bool   `json:"forbid_sort"`
	SeasonTitle   string `json:"season_title"`
	LongTitle     string `json:"long_title"`
	IndexTitle    string `json:"index_title"`
	CSource       string `json:"c_source"`
	MissionId     int    `json:"mission_id,omitempty"`
	UpFromV2      int    `json:"up_from_v2,omitempty"`
}

func (i *WaitWatchVideoResponse) getCode() int {
	return i.Code
}
func (i *WaitWatchVideoResponse) bindJSON(body []byte) error {
	return json.Unmarshal(body, i)
}

type WaitWatch struct {
}

func (w WaitWatch) getRequest(cookie *cookies.UserCookie) *http.Request {
	req, _ := http.NewRequest("GET", "https://api.bilibili.com/x/v2/history/toview", nil)
	req.Header.Add("Cookie", cookie.GetCookies())
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	return req
}
func (w WaitWatch) getResponse(cookie *cookies.UserCookie, useProxy bool) *WaitWatchVideoResponse {
	response, err := proxy.GetClient(useProxy).Do(w.getRequest(cookie))
	if err != nil {
		return nil
	}
	responseBody := new(WaitWatchVideoResponse)
	err = responseCodeCheck(response, responseBody, cookie)
	if err != nil {
		return nil
	}
	return responseBody
}
