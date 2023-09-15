package bilibili

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"videoDynamicAcquisition/utils"
)

type historyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Cursor struct {
			Max      int    `json:"max"`
			ViewAt   int    `json:"view_at"`
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
			Cover     string      `json:"cover"`
			Covers    interface{} `json:"covers"`
			Uri       string      `json:"uri"`
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
			Videos     int    `json:"videos"`
			AuthorName string `json:"author_name"`
			AuthorFace string `json:"author_face"`
			AuthorMid  int64  `json:"author_mid"`
			ViewAt     int64  `json:"view_at"`
			Progress   int    `json:"progress"`
			Badge      string `json:"badge"`
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
func (h *historyRequest) getRequest(max, viewAt int) *http.Request {
	url := "https://api.bilibili.com/x/web-interface/history/cursor"
	request, _ := http.NewRequest("GET", url, nil)
	q := request.URL.Query()
	q.Add("business", "archive")
	if max > 0 {
		q.Add("max", strconv.Itoa(max))
	}
	if viewAt > 0 {
		q.Add("view_at", strconv.Itoa(viewAt))
	}
	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", bilibiliCookies.cookies)
	return request
}

func (h *historyRequest) getResponse(max, viewAt int) *historyResponse {
	bilibiliCookies.flushCookies()
	if !bilibiliCookies.cookiesFail {
		return nil
	}
	request := h.getRequest(max, viewAt)
	response, _ := http.DefaultClient.Do(request)
	defer response.Body.Close()
	var responseData *historyResponse
	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		utils.ErrorLog.Println(string(body))
		return nil
	}
	if err != nil {
		utils.ErrorLog.Println("读取响应失败")
		utils.ErrorLog.Println(err.Error())
		return nil
	}
	responseData = new(historyResponse)
	err = json.Unmarshal(body, responseData)
	if err != nil {
		utils.ErrorLog.Println("解析响应失败")
		utils.ErrorLog.Println(err.Error())
		return nil
	}
	if responseData.Code == -101 {
		// cookies失效
		utils.ErrorLog.Println("cookies失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		return nil
	}
	if responseData.Code == -352 {
		utils.ErrorLog.Println("352错误，拒绝访问")
		return nil
	}
	if responseData.Code != 0 {
		utils.ErrorLog.Println("响应状态码错误", responseData.Code)
		utils.ErrorLog.Printf("%+v\n", responseData)
		return nil
	}

	return responseData
}