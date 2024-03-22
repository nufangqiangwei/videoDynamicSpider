package bilibili

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/models"
)

// api文档 https://github.com/SocialSisterYi/bilibili-API-collect/blob/f9ee5c3b99335af6bef0d9d902101c565b3bea00/docs/user/info.md
// api接口 https://api.bilibili.com/x/space/wbi/acc/info?mid=

type AuthorInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Mid         int    `json:"mid"`
		Name        string `json:"name"`
		Sex         string `json:"sex"`
		Face        string `json:"face"`
		FaceNft     int    `json:"face_nft"`
		FaceNftType int    `json:"face_nft_type"`
		Sign        string `json:"sign"`
		Rank        int    `json:"rank"`
		Level       int    `json:"level"`
		Jointime    int    `json:"jointime"`
		Moral       int    `json:"moral"`
		Silence     int    `json:"silence"`
		Coins       int    `json:"coins"`
		FansBadge   bool   `json:"fans_badge"`
		FansMedal   struct {
			Show  bool        `json:"show"`
			Wear  bool        `json:"wear"`
			Medal interface{} `json:"medal"`
		} `json:"fans_medal"`
		Official struct {
			Role  int    `json:"role"`
			Title string `json:"title"`
			Desc  string `json:"desc"`
			Type  int    `json:"type"`
		} `json:"official"`
		Vip struct {
			Type       int   `json:"type"`
			Status     int   `json:"status"`
			DueDate    int64 `json:"due_date"`
			VipPayType int   `json:"vip_pay_type"`
			ThemeType  int   `json:"theme_type"`
			Label      struct {
				Path                  string `json:"path"`
				Text                  string `json:"text"`
				LabelTheme            string `json:"label_theme"`
				TextColor             string `json:"text_color"`
				BgStyle               int    `json:"bg_style"`
				BgColor               string `json:"bg_color"`
				BorderColor           string `json:"border_color"`
				UseImgLabel           bool   `json:"use_img_label"`
				ImgLabelUriHans       string `json:"img_label_uri_hans"`
				ImgLabelUriHant       string `json:"img_label_uri_hant"`
				ImgLabelUriHansStatic string `json:"img_label_uri_hans_static"`
				ImgLabelUriHantStatic string `json:"img_label_uri_hant_static"`
			} `json:"label"`
			AvatarSubscript    int    `json:"avatar_subscript"`
			NicknameColor      string `json:"nickname_color"`
			Role               int    `json:"role"`
			AvatarSubscriptUrl string `json:"avatar_subscript_url"`
			TvVipStatus        int    `json:"tv_vip_status"`
			TvVipPayType       int    `json:"tv_vip_pay_type"`
			TvDueDate          int    `json:"tv_due_date"`
			AvatarIcon         struct {
				IconResource struct {
				} `json:"icon_resource"`
			} `json:"avatar_icon"`
		} `json:"vip"`
		Pendant struct {
			Pid               int    `json:"pid"`
			Name              string `json:"name"`
			Image             string `json:"image"`
			Expire            int    `json:"expire"`
			ImageEnhance      string `json:"image_enhance"`
			ImageEnhanceFrame string `json:"image_enhance_frame"`
			NPid              int    `json:"n_pid"`
		} `json:"pendant"`
		Nameplate struct {
			Nid        int    `json:"nid"`
			Name       string `json:"name"`
			Image      string `json:"image"`
			ImageSmall string `json:"image_small"`
			Level      string `json:"level"`
			Condition  string `json:"condition"`
		} `json:"nameplate"`
		UserHonourInfo struct {
			Mid               int           `json:"mid"`
			Colour            interface{}   `json:"colour"`
			Tags              []interface{} `json:"tags"`
			IsLatest100Honour int           `json:"is_latest_100honour"`
		} `json:"user_honour_info"`
		IsFollowed bool   `json:"is_followed"`
		TopPhoto   string `json:"top_photo"`
		Theme      struct {
		} `json:"theme"`
		SysNotice struct {
		} `json:"sys_notice"`
		LiveRoom struct {
			RoomStatus    int    `json:"roomStatus"`
			LiveStatus    int    `json:"liveStatus"`
			Url           string `json:"url"`
			Title         string `json:"title"`
			Cover         string `json:"cover"`
			Roomid        int    `json:"roomid"`
			RoundStatus   int    `json:"roundStatus"`
			BroadcastType int    `json:"broadcast_type"`
			WatchedShow   struct {
				Switch       bool   `json:"switch"`
				Num          int    `json:"num"`
				TextSmall    string `json:"text_small"`
				TextLarge    string `json:"text_large"`
				Icon         string `json:"icon"`
				IconLocation string `json:"icon_location"`
				IconWeb      string `json:"icon_web"`
			} `json:"watched_show"`
		} `json:"live_room"`
		Birthday string `json:"birthday"`
		School   struct {
			Name string `json:"name"`
		} `json:"school"`
		Profession struct {
			Name       string `json:"name"`
			Department string `json:"department"`
			Title      string `json:"title"`
			IsShow     int    `json:"is_show"`
		} `json:"profession"`
		Tags   interface{} `json:"tags"`
		Series struct {
			UserUpgradeStatus int  `json:"user_upgrade_status"`
			ShowUpgradeWindow bool `json:"show_upgrade_window"`
		} `json:"series"`
		IsSeniorMember int         `json:"is_senior_member"`
		McnInfo        interface{} `json:"mcn_info"`
		GaiaResType    int         `json:"gaia_res_type"`
		GaiaData       interface{} `json:"gaia_data"`
		IsRisk         bool        `json:"is_risk"`
		Elec           struct {
			ShowInfo struct {
				Show    bool   `json:"show"`
				State   int    `json:"state"`
				Title   string `json:"title"`
				Icon    string `json:"icon"`
				JumpUrl string `json:"jump_url"`
			} `json:"show_info"`
		} `json:"elec"`
		Contract struct {
			IsDisplay       bool `json:"is_display"`
			IsFollowDisplay bool `json:"is_follow_display"`
		} `json:"contract"`
		CertificateShow bool `json:"certificate_show"`
	} `json:"data"`
}

func (ai *AuthorInfoResponse) getCode() int {
	return ai.Code
}
func (ai *AuthorInfoResponse) bindJSON(body []byte) error {
	return json.Unmarshal(body, ai)
}
func (ai *AuthorInfoResponse) AccountName() string {
	return ai.Data.Name
}
func (ai *AuthorInfoResponse) GetAuthorModel() models.Author {
	return models.Author{
		WebSiteId:    webSiteId,
		AuthorWebUid: strconv.Itoa(ai.Data.Mid),
		AuthorName:   ai.Data.Name,
		Avatar:       ai.Data.Face,
		AuthorDesc:   ai.Data.Sign,
		CreateTime:   time.Now(),
	}
}

const authorInfoUrl = "https://api.bilibili.com/x/space/wbi/acc/info"

type authorInfo struct {
	userCookie *cookies.UserCookie
}

func (a authorInfo) getRequest(mid string) *http.Request {
	req, _ := http.NewRequest("GET", authorInfoUrl, nil)
	q := req.URL.Query()
	q.Add("mid", mid)
	q.Add("wts", strconv.FormatInt(time.Now().Unix(), 10))
	req.URL.RawQuery = wbiSignObj.getSing(q).Encode()
	req.Header.Add("Cookie", a.userCookie.GetCookies())
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	return req
}
func (a authorInfo) getResponse(mid string) *AuthorInfoResponse {
	a.userCookie.FlushCookies()
	if !a.userCookie.GetStatus() {
		return nil
	}
	response, err := http.DefaultClient.Do(a.getRequest(mid))
	if err != nil {
		return nil
	}
	responseBody := new(AuthorInfoResponse)
	err = responseCodeCheck(response, responseBody, a.userCookie)
	if err != nil {
		return nil
	}
	return responseBody

}

func getSelfInfo(mid, cookiesContext string) (*AuthorInfoResponse, error) {
	defaultUser := cookies.NewTemporaryUserCookie(webSiteName, cookiesContext)
	ai := authorInfo{
		userCookie: &defaultUser,
	}

	response := ai.getResponse(mid)
	if response == nil {
		return nil, errors.New("请求失败")
	}
	return response, nil
}
