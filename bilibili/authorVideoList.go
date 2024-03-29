package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
)

/*
https://github.com/SocialSisterYi/bilibili-API-collect/blob/f9ee5c3b99335af6bef0d9d902101c565b3bea00/docs/user/space.md?plain=1#L2233
{
	"GET": {
		"scheme": "https",
		"host": "api.bilibili.com",
		"filename": "/x/space/wbi/arc/search",
		"query": {
			"mid": "11352614",
			"ps": "50",
			"tid": "4",
			"special_type": "",
			"pn": "1", // 页码
			"keyword": "",
			"order": "pubdate",
			"platform": "web",
			"web_location": "1550101",
			"order_avoided": "true",
			"w_rid": "bb2818a9f8cf0a38e55d6ad1d856cec0",
			"wts": "1693900786"
		},
		"remote": {
			"地址": "127.0.0.1:10809"
		}
	}
}
https://api.bilibili.com/x/space/wbi/arc/search?mid=11352614&ps=30&tid=4&special_type=&pn=1&keyword=&order=pubdate&platform=web&web_location=1550101&order_avoided=true&w_rid=5f31001692c32ed3c8f2690b93f0b086&wts=1693906320
https://api.bilibili.com/x/space/wbi/arc/search?mid=11352614&ps=30&tid=4&special_type=&pn=2&keyword=&order=pubdate&platform=web&web_location=1550101&order_avoided=true&w_rid=8d5962c646281e5e7247888dc22251c9&wts=1693906362
*/

type VideoListPageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		List struct {
			Tlist struct {
				Field1 struct {
					Tid   int    `json:"tid"`
					Count int    `json:"count"`
					Name  string `json:"name"`
				} `json:"155"`
				Field2 struct {
					Tid   int    `json:"tid"`
					Count int    `json:"count"`
					Name  string `json:"name"`
				} `json:"160"`
				Field3 struct {
					Tid   int    `json:"tid"`
					Count int    `json:"count"`
					Name  string `json:"name"`
				} `json:"217"`
				Field4 struct {
					Tid   int    `json:"tid"`
					Count int    `json:"count"`
					Name  string `json:"name"`
				} `json:"4"`
			} `json:"tlist"`
			Vlist []VideoList   `json:"vlist"`
			Slist []interface{} `json:"slist"`
		} `json:"list"`
		Page struct {
			Pn    int `json:"pn"`
			Ps    int `json:"ps"`
			Count int `json:"count"`
		} `json:"page"`
		EpisodicButton struct {
			Text string `json:"text"`
			Uri  string `json:"uri"`
		} `json:"episodic_button"`
		IsRisk      bool        `json:"is_risk"`
		GaiaResType int         `json:"gaia_res_type"`
		GaiaData    interface{} `json:"gaia_data"`
	} `json:"data"`
}
type VideoList struct {
	Comment        int         `json:"comment"`
	Typeid         int         `json:"typeid"`
	Play           interface{} `json:"play"`
	Pic            string      `json:"pic"`
	Subtitle       string      `json:"subtitle"`
	Description    string      `json:"description"`
	Copyright      string      `json:"copyright"`
	Title          string      `json:"title"`
	Review         int         `json:"review"`
	Author         string      `json:"author"`
	Mid            int         `json:"mid"`
	Created        int64       `json:"created"`
	Length         string      `json:"length"`
	VideoReview    int         `json:"video_review"`
	Aid            int         `json:"aid"`
	Bvid           string      `json:"bvid"`
	HideClick      bool        `json:"hide_click"`
	IsPay          int         `json:"is_pay"`
	IsUnionVideo   int         `json:"is_union_video"` // 联合投稿
	IsSteinsGate   int         `json:"is_steins_gate"`
	IsLivePlayback int         `json:"is_live_playback"`
	Meta           *struct {
		Id        int    `json:"id"`
		Title     string `json:"title"`
		Cover     string `json:"cover"`
		Mid       int    `json:"mid"`
		Intro     string `json:"intro"`
		SignState int    `json:"sign_state"`
		Attribute int    `json:"attribute"`
		Stat      struct {
			SeasonId int `json:"season_id"`
			View     int `json:"view"`
			Danmaku  int `json:"danmaku"`
			Reply    int `json:"reply"`
			Favorite int `json:"favorite"`
			Coin     int `json:"coin"`
			Share    int `json:"share"`
			Like     int `json:"like"`
			Mtime    int `json:"mtime"`
			Vt       int `json:"vt"`
			Vv       int `json:"vv"`
		} `json:"stat"`
		EpCount  int `json:"ep_count"`
		FirstAid int `json:"first_aid"`
		Ptime    int `json:"ptime"`
		EpNum    int `json:"ep_num"`
	} `json:"meta"`
	IsAvoided     int    `json:"is_avoided"`
	Attribute     int    `json:"attribute"`
	IsChargingArc bool   `json:"is_charging_arc"`
	Vt            int    `json:"vt"`
	EnableVt      int    `json:"enable_vt"`
	VtDisplay     string `json:"vt_display"`
}
type videoListPage struct {
	userCookie *cookies.UserCookie
}

func (v *videoListPage) getRequest(mid string, pageIndex int) *http.Request {
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/space/wbi/arc/search", nil)
	q := request.URL.Query()
	q.Add("mid", mid)
	q.Add("ps", "50")
	q.Add("platform", "web")
	q.Add("dm_img_list", "[]")
	q.Add("dm_img_str", "")
	q.Add("dm_cover_img_str", "")
	q.Add("dm_img_inter", "[]")
	q.Add("pn", strconv.Itoa(pageIndex))

	request.URL.RawQuery = wbiSignObj.getSing(q).Encode()
	request.Header.Add("Cookie", v.userCookie.GetCookies())
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edg/116.0.1938.69")
	//utils.Info.Println(request.URL.String())
	return request
}

func (v *videoListPage) getResponse(mid string, pageIndex int) *VideoListPageResponse {
	v.userCookie.FlushCookies()
	if !v.userCookie.GetStatus() {
		return nil
	}
	response, err := http.DefaultClient.Do(v.getRequest(mid, pageIndex))

	if err != nil {
		log.ErrorLog.Println(err.Error())
		return nil
	}
	responseBody := new(VideoListPageResponse)
	err = responseCodeCheck(response, responseBody, v.userCookie)
	if err != nil {
		return nil
	}
	return responseBody
}

func GetAuthorAllVideoListByByte(uid string, pageIndex int) ([]byte, error, string) {
	defaultUser := getUser()
	v := videoListPage{
		userCookie: defaultUser,
	}
	response, err := http.DefaultClient.Do(v.getRequest(uid, pageIndex))

	if err != nil {
		return nil, err, response.Request.URL.String()
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return data, errors.New(fmt.Sprintf("StatusCode错误%d", response.StatusCode)), response.Request.URL.String()
	}
	return data, err, response.Request.URL.String()
}
func GetAuthorAllVideoList(uid string, pageIndex int) *VideoListPageResponse {
	defaultUser := getUser()
	v := videoListPage{
		userCookie: defaultUser,
	}
	return v.getResponse(uid, pageIndex)
}

func (r *VideoListPageResponse) getCode() int {
	return r.Code
}
func (r *VideoListPageResponse) bindJSON(data []byte) error {
	return json.Unmarshal(data, r)
}
