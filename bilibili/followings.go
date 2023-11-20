package bilibili

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

// https://github.com/SocialSisterYi/bilibili-API-collect/blob/d6d17871459370883f4fd105161df3ce8db31b9d/docs/user/relation.md
type followingsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		List      []FollowingUP `json:"list"`
		ReVersion int           `json:"re_version"`
		Total     int           `json:"total"`
	} `json:"data"`
}
type FollowingUP struct {
	Mid          int64       `json:"mid"`
	Attribute    int         `json:"attribute"`
	Mtime        int64       `json:"mtime"` // 关注对方时间
	Tag          interface{} `json:"tag"`
	Special      int         `json:"special"`
	ContractInfo struct {
	} `json:"contract_info"`
	Uname          string `json:"uname"`
	Face           string `json:"face"`
	Sign           string `json:"sign"`
	FaceNft        int    `json:"face_nft"`
	OfficialVerify struct {
		Type int    `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_verify"`
	Vip struct {
		VipType       int    `json:"vipType"`
		VipDueDate    int64  `json:"vipDueDate"`
		DueRemark     string `json:"dueRemark"`
		AccessStatus  int    `json:"accessStatus"`
		VipStatus     int    `json:"vipStatus"`
		VipStatusWarn string `json:"vipStatusWarn"`
		ThemeType     int    `json:"themeType"`
		Label         struct {
			Path        string `json:"path"`
			Text        string `json:"text"`
			LabelTheme  string `json:"label_theme"`
			TextColor   string `json:"text_color"`
			BgStyle     int    `json:"bg_style"`
			BgColor     string `json:"bg_color"`
			BorderColor string `json:"border_color"`
		} `json:"label"`
		AvatarSubscript    int    `json:"avatar_subscript"`
		NicknameColor      string `json:"nickname_color"`
		AvatarSubscriptUrl string `json:"avatar_subscript_url"`
	} `json:"vip"`
	NftIcon   string `json:"nft_icon"`
	RecReason string `json:"rec_reason"`
	TrackId   string `json:"track_id"`
}

type followings struct {
	pageNumber int
}

func (f *followings) getRequest() *http.Request {
	// https://api.bilibili.com/x/relation/followings?vmid=10932398&pn=1&ps=20&order=desc&order_type=&gaia_source=main_web
	request, _ := http.NewRequest("GET", followingsBseUrl, nil)
	request.Header.Add("Cookie", bilibiliCookies.cookies)
	//request.Header.Add("User-Agent", "Mozilla/5.0")
	q := request.URL.Query()
	q.Add("vmid", "10932398")
	q.Add("pn", strconv.Itoa(f.pageNumber))
	q.Add("ps", "20")
	q.Add("order", "desc")
	q.Add("order_type", "")
	q.Add("gaia_source", "main_web")
	request.URL.RawQuery = q.Encode()
	return request
}

func (f *followings) getResponse(retriesNumber int) (followingsResponseBody followingsResponse) {
	if retriesNumber > 2 {
		utils.ErrorLog.Println("重试次数过多")
		return
	}
	bilibiliCookies.flushCookies()
	request := f.getRequest()
	//utils.Info.Println("请求地址", request.URL.String())
	client := &http.Client{}
	response, err := client.Do(request)
	followingsResponseBody = followingsResponse{}
	if err != nil {
		utils.ErrorLog.Println("请求失败")
		utils.ErrorLog.Println(err.Error())
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		utils.ErrorLog.Println("读取响应失败")
		utils.ErrorLog.Println(err.Error())
		return
	}
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		utils.ErrorLog.Println(string(body))
		return
	}

	err = json.Unmarshal(body, &followingsResponseBody)
	if err != nil {
		utils.ErrorLog.Println("解析响应失败")
		utils.ErrorLog.Println(err.Error())
		return
	}
	if followingsResponseBody.Code == -101 {
		// cookies失效
		utils.ErrorLog.Println("cookies失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		if bilibiliCookies.cookiesFail {
			f.getResponse(retriesNumber + 1)
		} else {
			utils.ErrorLog.Println("cookies失效，请更新cookies文件")
			return
		}
	}
	if followingsResponseBody.Code != 0 {
		utils.ErrorLog.Println("响应状态码错误", followingsResponseBody.Code)
		utils.ErrorLog.Printf("%+v", followingsResponseBody)
		return
	}
	//utils.Info.Println("关注列表获取成功")
	return followingsResponseBody

}

func (f *followings) getFollowings(webSiteId int64) (result []models.Author) {
	f.pageNumber = 1
	result = make([]models.Author, 0)
	response := f.getResponse(0)
	followingNumber := response.Data.Total
	utils.Info.Println("关注总数", followingNumber)
	for f.pageNumber = 1; f.pageNumber <= followingNumber/50+1; f.pageNumber++ {
		for _, info := range response.Data.List {
			result = append(result, models.Author{
				WebSiteId:    webSiteId,
				AuthorWebUid: strconv.FormatInt(info.Mid, 10),
				AuthorName:   info.Uname,
				Avatar:       info.Face,
				AuthorDesc:   info.Sign,
			})
		}
		time.Sleep(time.Second * 10)
		response = f.getResponse(0)
	}

	return
}
