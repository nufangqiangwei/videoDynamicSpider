package bilibili

import (
	"encoding/json"
	"net/http"
	"strconv"
	"videoDynamicAcquisition/utils"
)

// https://github.com/SocialSisterYi/bilibili-API-collect/blob/ffa25ba78dc8f4ed8624f11e3b6f404cb799674f/docs/fav/list.md api文档

type collectListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Count int `json:"count"`
		List  []struct {
			Id         int64  `json:"id"`
			Fid        int    `json:"fid"`
			Mid        int    `json:"mid"`
			Attr       int    `json:"attr"`
			Title      string `json:"title"`
			FavState   int    `json:"fav_state"`
			MediaCount int    `json:"media_count"`
		} `json:"list"`
		Season interface{} `json:"season"`
	} `json:"data"`
}

type subscriptionListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Count int `json:"count"`
		List  []struct {
			Id    int64  `json:"id"`
			Fid   int    `json:"fid"`
			Mid   int64  `json:"mid"`
			Attr  int    `json:"attr"`
			Title string `json:"title"`
			Cover string `json:"cover"`
			Upper struct {
				Mid  int64  `json:"mid"`
				Name string `json:"name"`
				Face string `json:"face"`
			} `json:"upper"`
			CoverType  int    `json:"cover_type"`
			Intro      string `json:"intro"`
			Ctime      int    `json:"ctime"`
			Mtime      int    `json:"mtime"`
			State      int    `json:"state"`
			FavState   int    `json:"fav_state"`
			MediaCount int    `json:"media_count"`
			ViewCount  int    `json:"view_count"`
			Vt         int    `json:"vt"`
			PlaySwitch int    `json:"play_switch"`
			Type       int    `json:"type"`
			Link       string `json:"link"`
			Bvid       string `json:"bvid"`
		} `json:"list"`
		HasMore bool `json:"has_more"`
	} `json:"data"`
}

type CollectAllVideoListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    []struct {
		Id   int    `json:"id"`
		Type int    `json:"type"`
		BvId string `json:"bv_id"`
		Bvid string `json:"bvid"`
	} `json:"data"`
}

type seasonAllVideoDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Info struct {
			Id         int    `json:"id"`
			SeasonType int    `json:"season_type"`
			Title      string `json:"title"`
			Cover      string `json:"cover"`
			Upper      struct {
				Mid  int    `json:"mid"`
				Name string `json:"name"`
			} `json:"upper"`
			CntInfo struct {
				Collect int `json:"collect"`
				Play    int `json:"play"`
				Danmaku int `json:"danmaku"`
				Vt      int `json:"vt"`
			} `json:"cnt_info"`
			MediaCount int    `json:"media_count"` // 视频总数
			Intro      string `json:"intro"`
			EnableVt   int    `json:"enable_vt"`
		} `json:"info"`
		Medias []SeasonAllVideoDetailInfo `json:"medias"`
	} `json:"data"`
}
type SeasonAllVideoDetailInfo struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Duration int    `json:"duration"`
	Pubtime  int    `json:"pubtime"`
	Bvid     string `json:"bvid"`
	Upper    struct {
		Mid  int    `json:"mid"`
		Name string `json:"name"`
	} `json:"upper"`
	CntInfo struct {
		Collect int `json:"collect"`
		Play    int `json:"play"`
		Danmaku int `json:"danmaku"`
		Vt      int `json:"vt"`
	} `json:"cnt_info"`
	EnableVt  int    `json:"enable_vt"`
	VtDisplay string `json:"vt_display"`
}

type collectVideoDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Info struct {
			Id    int    `json:"id"`
			Fid   int    `json:"fid"`
			Mid   int    `json:"mid"`
			Attr  int    `json:"attr"`
			Title string `json:"title"`
			Cover string `json:"cover"`
			Upper struct {
				Mid       int    `json:"mid"`
				Name      string `json:"name"`
				Face      string `json:"face"`
				Followed  bool   `json:"followed"`
				VipType   int    `json:"vip_type"`
				VipStatue int    `json:"vip_statue"`
			} `json:"upper"`
			CoverType int `json:"cover_type"`
			CntInfo   struct {
				Collect int `json:"collect"`
				Play    int `json:"play"`
				ThumbUp int `json:"thumb_up"`
				Share   int `json:"share"`
			} `json:"cnt_info"`
			Type       int    `json:"type"`
			Intro      string `json:"intro"`
			Ctime      int    `json:"ctime"`
			Mtime      int    `json:"mtime"`
			State      int    `json:"state"`
			FavState   int    `json:"fav_state"`
			LikeState  int    `json:"like_state"`
			MediaCount int    `json:"media_count"` // 视频总数
		} `json:"info"`
		Medias  []CollectVideoDetailInfo `json:"medias"`
		HasMore bool                     `json:"has_more"`
	} `json:"data"`
}
type CollectVideoDetailInfo struct {
	Id       int    `json:"id"`
	Type     int    `json:"type"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Intro    string `json:"intro"`
	Page     int    `json:"page"`
	Duration int    `json:"duration"`
	Upper    struct {
		Mid  int    `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"upper"`
	Attr    int `json:"attr"`
	CntInfo struct {
		Collect int `json:"collect"`
		Play    int `json:"play"`
		Danmaku int `json:"danmaku"`
	} `json:"cnt_info"`
	Link    string      `json:"link"`
	Ctime   int64       `json:"ctime"`
	Pubtime int64       `json:"pubtime"`
	FavTime int64       `json:"fav_time"`
	BvId    string      `json:"bv_id"`
	Bvid    string      `json:"bvid"`
	Season  interface{} `json:"season"`
}

// https://api.bilibili.com/x/v3/fav/folder/created/list-all?up_mid=10932398 获取收藏夹列表
func getCollectList(mid string) *collectListResponse {
	bilibiliCookies.flushCookies()
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/v3/fav/folder/created/list-all", nil)
	q := request.URL.Query()
	q.Add("up_mid", mid)
	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", bilibiliCookies.cookies)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		utils.ErrorLog.Println("请求错误", err.Error())
		return nil
	}
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		return nil
	}
	var result collectListResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		utils.ErrorLog.Println("解析错误", err.Error())
		return nil
	}
	if result.Code == -101 {
		utils.ErrorLog.Println("cookie失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		return nil
	}
	return &result
}

// https://api.bilibili.com/x/v3/fav/folder/collected/list?pn=1&ps=20&up_mid=10932398&platform=web 获取收藏和订阅列表
func subscriptionList(mid string) *subscriptionListResponse {
	bilibiliCookies.flushCookies()
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/v3/fav/folder/collected/list", nil)
	q := request.URL.Query()
	q.Add("up_mid", mid)
	q.Add("pn", "1")
	q.Add("ps", "20")
	q.Add("platform", "web")
	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", bilibiliCookies.cookies)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		utils.ErrorLog.Println("请求错误", err.Error())
		return nil
	}
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		return nil
	}
	var result subscriptionListResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		utils.ErrorLog.Println("解析错误", err.Error())
		return nil
	}
	if result.Code == -101 {
		utils.ErrorLog.Println("cookie失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		return nil
	}
	return &result
}

// GetCollectVideoList 只能获取个人收藏夹的视频列表
// https://api.bilibili.com/x/v3/fav/resource/ids?media_id=55899098&platform=web
func GetCollectVideoList(id int64) *CollectAllVideoListResponse {
	bilibiliCookies.flushCookies()
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/v3/fav/resource/ids", nil)
	q := request.URL.Query()
	q.Add("media_id", strconv.FormatInt(id, 10))
	q.Add("platform", "web")
	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", bilibiliCookies.cookies)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		utils.ErrorLog.Println("请求错误", err.Error())
		return nil
	}
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		return nil
	}
	var result CollectAllVideoListResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		utils.ErrorLog.Println("解析错误", err.Error())
		return nil
	}
	if result.Code == -101 {
		utils.ErrorLog.Println("cookie失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		return nil
	}
	return &result
}

// 只能获取个人收藏夹的视频列表
// https://api.bilibili.com/x/v3/fav/resource/list?media_id=55899098&pn=1&ps=20&keyword=&order=mtime&type=0&tid=0&platform=web
func getCollectVideoInfo(collectId int64, page int) *collectVideoDetailResponse {
	bilibiliCookies.flushCookies()
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/v3/fav/resource/list", nil)
	q := request.URL.Query()
	q.Add("media_id", strconv.FormatInt(collectId, 10))
	q.Add("pn", strconv.Itoa(page))
	q.Add("ps", "20")
	q.Add("keyword", "")
	q.Add("order", "mtime")
	q.Add("type", "0")
	q.Add("tid", "0")
	q.Add("platform", "web")
	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", bilibiliCookies.cookies)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		utils.ErrorLog.Println("请求错误", err.Error())
		return nil
	}
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		return nil
	}
	var result collectVideoDetailResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		utils.ErrorLog.Println("解析错误", err.Error())
		return nil
	}
	if result.Code == -101 {
		utils.ErrorLog.Println("cookie失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		return nil
	}
	return &result

}

// https://api.bilibili.com/x/space/fav/season/list?season_id=1057928&pn=1&ps=20
func getSeasonVideoInfo(collectId int64, page int) *seasonAllVideoDetailResponse {
	bilibiliCookies.flushCookies()
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/space/fav/season/list", nil)
	q := request.URL.Query()
	q.Add("season_id", strconv.FormatInt(collectId, 10))
	q.Add("pn", strconv.Itoa(page))
	q.Add("ps", "20")
	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", bilibiliCookies.cookies)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		utils.ErrorLog.Println("请求错误", err.Error())
		return nil
	}
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		return nil
	}
	var result seasonAllVideoDetailResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		utils.ErrorLog.Println("解析错误", err.Error())
		return nil
	}
	if result.Code == -101 {
		utils.ErrorLog.Println("cookie失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		return nil
	}
	return &result
}
