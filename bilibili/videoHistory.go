package bilibili

import (
	"encoding/json"
	"net/http"
	"strconv"
	"videoDynamicAcquisition/utils"
)

// HistoryResponse https://github.com/SocialSisterYi/bilibili-API-collect/blob/ffa25ba78dc8f4ed8624f11e3b6f404cb799674f/docs/history%26toview/history.md
type HistoryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Cursor struct {
			Max      int    `json:"max"`
			ViewAt   int64  `json:"view_at"`
			Business string `json:"business"`
			Ps       int    `json:"ps"`
		} `json:"cursor"`
		Tab []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"tab"`
		List []struct {
			Title     string      `json:"title"`
			LongTitle string      `json:"long_title"`
			Cover     string      `json:"cover"`  // 用于专栏以外的条目
			Covers    interface{} `json:"covers"` // 仅用于专栏.有效时：array 无效时：null
			Uri       string      `json:"uri"`    // 仅用于剧集和直播
			History   struct {
				Oid      int    `json:"oid"`
				Epid     int    `json:"epid"`
				Bvid     string `json:"bvid"`
				Page     int    `json:"page"`
				Cid      int    `json:"cid"`
				Part     string `json:"part"`
				Business string `json:"business"`
				Dt       int    `json:"dt"`
			} `json:"history"`
			Videos     int    `json:"videos"` // 仅用于稿件视频
			AuthorName string `json:"author_name"`
			AuthorFace string `json:"author_face"`
			AuthorMid  int64  `json:"author_mid"`
			ViewAt     int64  `json:"view_at"` // 时间戳
			Progress   int    `json:"progress"`
			Badge      string `json:"badge"` // 空字符串时是稿件视频 / 剧集 / 笔记 / 纪录片 / 专栏 / 国创 / 番剧
			ShowTitle  string `json:"show_title"`
			Duration   int    `json:"duration"`
			Current    string `json:"current"`
			Total      int    `json:"total"`
			NewDesc    string `json:"new_desc"`
			IsFinish   int    `json:"is_finish"`
			IsFav      int    `json:"is_fav"`
			Kid        int64  `json:"kid"`
			TagName    string `json:"tag_name"`
			LiveStatus int    `json:"live_status"`
		} `json:"list"`
	} `json:"data"`
}

type historyRequest struct{}

//
func (h *historyRequest) getRequest(max int, viewAt int64, business string) *http.Request {
	url := "https://api.bilibili.com/x/web-interface/history/cursor"
	request, _ := http.NewRequest("GET", url, nil)
	q := request.URL.Query()
	q.Add("business", business)
	if max > 0 {
		q.Add("max", strconv.Itoa(max))
	}
	if viewAt > 0 {
		q.Add("view_at", strconv.FormatInt(viewAt, 10))
	}
	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", biliCookiesManager.getUser(DefaultCookies).cookies)
	return request
}

func (h *historyRequest) getResponse(max int, viewAt int64, business string) *HistoryResponse {
	biliCookiesManager.getUser(DefaultCookies).flushCookies()
	if !biliCookiesManager.getUser(DefaultCookies).cookiesFail {
		return nil
	}
	request := h.getRequest(max, viewAt, business)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		utils.ErrorLog.Printf("获取历史记录错误:%s", err.Error())
		return nil
	}
	result := new(HistoryResponse)
	err = responseCodeCheck(response, result)
	if err != nil {
		return nil
	}
	return result
}

// historyResponse实现responseCheck接口
func (h *HistoryResponse) getCode() int {
	return h.Code
}
func (h *HistoryResponse) bindJSON(body []byte) error {
	return json.Unmarshal(body, h)
}
func (h *HistoryResponse) BingJSON(body []byte) error {
	return json.Unmarshal(body, h)
}
