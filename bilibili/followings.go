package bilibili

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
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
	pageNumber  int
	userCookies *cookies.UserCookie
}

func (f *followings) getRequest() *http.Request {
	// https://api.bilibili.com/x/relation/followings?vmid=10932398&pn=1&ps=20&order=desc&order_type=&gaia_source=main_web
	request, _ := http.NewRequest("GET", followingsBseUrl, nil)
	request.Header.Add("Cookie", f.userCookies.GetCookies())
	//request.Header.Add("User-Agent", "Mozilla/5.0")
	q := request.URL.Query()
	q.Add("vmid", f.userCookies.GetCookiesKeyValue("DedeUserID"))
	q.Add("pn", strconv.Itoa(f.pageNumber))
	q.Add("ps", "20")
	q.Add("order", "desc")
	q.Add("order_type", "")
	q.Add("gaia_source", "main_web")
	request.URL.RawQuery = q.Encode()
	return request
}

func (f *followings) getResponse(retriesNumber int) *followingsResponse {
	if retriesNumber > 2 {
		log.ErrorLog.Println("重试次数过多")
		return nil
	}
	f.userCookies.FlushCookies()
	request := f.getRequest()
	//utils.Info.Println("请求地址", request.URL.String())
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.ErrorLog.Println("请求失败")
		log.ErrorLog.Println(err.Error())
		return nil
	}
	result := new(followingsResponse)
	err = responseCodeCheck(response, result, f.userCookies)
	if err != nil {
		return nil
	}
	return result

}

func (f *followings) getFollowings(webSiteId int64) (result []models.Author) {
	f.pageNumber = 1
	result = make([]models.Author, 0)
	response := f.getResponse(0)
	if response == nil {
		response = &followingsResponse{}
	}
	followingNumber := response.Data.Total
	log.Info.Println("关注总数", followingNumber)
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
		if response == nil {
			response = &followingsResponse{}
		}
	}

	return
}

// followingsResponse实现responseCheck接口
func (f *followingsResponse) getCode() int {
	return f.Code
}
func (f *followingsResponse) bindJSON(body []byte) error {
	return json.Unmarshal(body, f)
}
