package bilibili

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"videoDynamicAcquisition/utils"
)

// https://socialsisteryi.github.io/bilibili-API-collect/docs/video/info.html
// https://api.bilibili.com/x/web-interface/view/detail?bvid=BV117411r7R1
type VideoDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		View struct {
			Bvid      string `json:"bvid"`
			Aid       int    `json:"aid"`
			Videos    int    `json:"videos"`
			Tid       int    `json:"tid"`
			Tname     string `json:"tname"`
			Copyright int    `json:"copyright"`
			Pic       string `json:"pic"`
			Title     string `json:"title"`
			Pubdate   int    `json:"pubdate"`
			Ctime     int64  `json:"ctime"`
			Desc      string `json:"desc"`
			DescV2    []struct {
				RawText string `json:"raw_text"`
				Type    int    `json:"type"`
				BizId   int    `json:"biz_id"`
			} `json:"desc_v2"`
			State     int `json:"state"`
			Duration  int `json:"duration"`
			MissionId int `json:"mission_id"`
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
				CleanMode     int `json:"clean_mode"`
				IsSteinGate   int `json:"is_stein_gate"`
				Is360         int `json:"is_360"`
				NoShare       int `json:"no_share"`
				ArcPay        int `json:"arc_pay"`
				FreeWatch     int `json:"free_watch"`
			} `json:"rights"`
			Owner struct {
				Mid  int    `json:"mid"`
				Name string `json:"name"`
				Face string `json:"face"`
			} `json:"owner"` // 上传者
			Stat struct {
				Aid        int64  `json:"aid"`
				View       int64  `json:"view"`
				Danmaku    int64  `json:"danmaku"`
				Reply      int64  `json:"reply"`
				Favorite   int64  `json:"favorite"`
				Coin       int64  `json:"coin"`
				Share      int64  `json:"share"`
				NowRank    int64  `json:"now_rank"`
				HisRank    int64  `json:"his_rank"`
				Like       int64  `json:"like"`
				Dislike    int64  `json:"dislike"`
				Evaluation string `json:"evaluation"`
				ArgueMsg   string `json:"argue_msg"`
				Vt         int    `json:"vt"`
			} `json:"stat"`
			Dynamic   string `json:"dynamic"`
			Cid       int    `json:"cid"`
			Dimension struct {
				Width  int `json:"width"`
				Height int `json:"height"`
				Rotate int `json:"rotate"`
			} `json:"dimension"`
			Premiere           interface{} `json:"premiere"`
			TeenageMode        int         `json:"teenage_mode"`
			IsChargeableSeason bool        `json:"is_chargeable_season"`
			IsStory            bool        `json:"is_story"`
			IsUpowerExclusive  bool        `json:"is_upower_exclusive"`
			IsUpowerPlay       bool        `json:"is_upower_play"`
			EnableVt           int         `json:"enable_vt"`
			VtDisplay          string      `json:"vt_display"`
			NoCache            bool        `json:"no_cache"`
			Pages              []struct {
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
			} `json:"pages"`
			Subtitle struct {
				AllowSubmit bool `json:"allow_submit"`
				List        []struct {
					Id          int64  `json:"id"`
					Lan         string `json:"lan"`
					LanDoc      string `json:"lan_doc"`
					IsLock      bool   `json:"is_lock"`
					SubtitleUrl string `json:"subtitle_url"`
					Type        int    `json:"type"`
					IdStr       string `json:"id_str"`
					AiType      int    `json:"ai_type"`
					AiStatus    int    `json:"ai_status"`
					Author      struct {
						Mid            int    `json:"mid"`
						Name           string `json:"name"`
						Sex            string `json:"sex"`
						Face           string `json:"face"`
						Sign           string `json:"sign"`
						Rank           int    `json:"rank"`
						Birthday       int    `json:"birthday"`
						IsFakeAccount  int    `json:"is_fake_account"`
						IsDeleted      int    `json:"is_deleted"`
						InRegAudit     int    `json:"in_reg_audit"`
						IsSeniorMember int    `json:"is_senior_member"`
					} `json:"author"`
				} `json:"list"`
			} `json:"subtitle"`
			Staff []struct {
				Mid   int    `json:"mid"`
				Title string `json:"title"`
				Name  string `json:"name"`
				Face  string `json:"face"`
				Vip   struct {
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
				} `json:"vip"`
				Official struct {
					Role  int    `json:"role"`
					Title string `json:"title"`
					Desc  string `json:"desc"`
					Type  int    `json:"type"`
				} `json:"official"`
				Follower   uint64 `json:"follower"`
				LabelStyle int    `json:"label_style"`
			} `json:"staff"` // 联合投稿人
			IsSeasonDisplay bool `json:"is_season_display"`
			UserGarb        struct {
				UrlImageAniCut string `json:"url_image_ani_cut"`
			} `json:"user_garb"`
			HonorReply struct {
				Honor []struct {
					Aid                int    `json:"aid"`
					Type               int    `json:"type"`
					Desc               string `json:"desc"`
					WeeklyRecommendNum int    `json:"weekly_recommend_num"`
				} `json:"honor"`
			} `json:"honor_reply"`
			LikeIcon          string `json:"like_icon"`
			NeedJumpBv        bool   `json:"need_jump_bv"`
			DisableShowUpInfo bool   `json:"disable_show_up_info"`
		} `json:"View"`
		Card struct {
			Card struct {
				Mid         string        `json:"mid"`
				Name        string        `json:"name"`
				Approve     bool          `json:"approve"`
				Sex         string        `json:"sex"`
				Rank        string        `json:"rank"`
				Face        string        `json:"face"`
				FaceNft     int           `json:"face_nft"`
				FaceNftType int           `json:"face_nft_type"`
				DisplayRank string        `json:"DisplayRank"`
				Regtime     int           `json:"regtime"`
				Spacesta    int           `json:"spacesta"`
				Birthday    string        `json:"birthday"`
				Place       string        `json:"place"`
				Description string        `json:"description"`
				Article     int           `json:"article"`
				Attentions  []interface{} `json:"attentions"`
				Fans        uint64        `json:"fans"`
				Friend      int           `json:"friend"`
				Attention   int           `json:"attention"`
				Sign        string        `json:"sign"`
				LevelInfo   struct {
					CurrentLevel int `json:"current_level"`
					CurrentMin   int `json:"current_min"`
					CurrentExp   int `json:"current_exp"`
					NextExp      int `json:"next_exp"`
				} `json:"level_info"`
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
				Official struct {
					Role  int    `json:"role"`
					Title string `json:"title"`
					Desc  string `json:"desc"`
					Type  int    `json:"type"`
				} `json:"Official"`
				OfficialVerify struct {
					Type int    `json:"type"`
					Desc string `json:"desc"`
				} `json:"official_verify"`
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
					VipType            int    `json:"vipType"`
					VipStatus          int    `json:"vipStatus"`
				} `json:"vip"`
				IsSeniorMember int `json:"is_senior_member"`
			} `json:"card"`
			Space struct {
				SImg string `json:"s_img"`
				LImg string `json:"l_img"`
			} `json:"space"`
			Following    bool `json:"following"`
			ArchiveCount int  `json:"archive_count"`
			ArticleCount int  `json:"article_count"`
			Follower     int  `json:"follower"`
			LikeNum      int  `json:"like_num"`
		} `json:"Card"`
		Tags []struct {
			TagId        int64  `json:"tag_id"`
			TagName      string `json:"tag_name"`
			Cover        string `json:"cover"`
			HeadCover    string `json:"head_cover"`
			Content      string `json:"content"`
			ShortContent string `json:"short_content"`
			Type         int    `json:"type"`
			State        int    `json:"state"`
			Ctime        int    `json:"ctime"`
			Count        struct {
				View  int `json:"view"`
				Use   int `json:"use"`
				Atten int `json:"atten"`
			} `json:"count"`
			IsAtten         int    `json:"is_atten"`
			Likes           int    `json:"likes"`
			Hates           int    `json:"hates"`
			Attribute       int    `json:"attribute"`
			Liked           int    `json:"liked"`
			Hated           int    `json:"hated"`
			ExtraAttr       int    `json:"extra_attr"`
			MusicId         string `json:"music_id"`
			TagType         string `json:"tag_type"`
			IsActivity      bool   `json:"is_activity"`
			Color           string `json:"color"`
			Alpha           int    `json:"alpha"`
			IsSeason        bool   `json:"is_season"`
			SubscribedCount int    `json:"subscribed_count"`
			ArchiveCount    string `json:"archive_count"`
			FeaturedCount   int    `json:"featured_count"`
			JumpUrl         string `json:"jump_url"`
		} `json:"Tags"`
		Reply struct {
			Page    interface{} `json:"page"`
			Replies []struct {
				Rpid       int         `json:"rpid"`
				Oid        int         `json:"oid"`
				Type       int         `json:"type"`
				Mid        int         `json:"mid"`
				Root       int         `json:"root"`
				Parent     int         `json:"parent"`
				Dialog     int         `json:"dialog"`
				Count      int         `json:"count"`
				Rcount     int         `json:"rcount"`
				State      int         `json:"state"`
				Fansgrade  int         `json:"fansgrade"`
				Attr       int         `json:"attr"`
				Ctime      int         `json:"ctime"`
				Like       int         `json:"like"`
				Action     int         `json:"action"`
				Content    interface{} `json:"content"`
				Replies    interface{} `json:"replies"`
				Assist     int         `json:"assist"`
				ShowFollow bool        `json:"show_follow"`
			} `json:"replies"`
		} `json:"Reply"`
		Related []struct {
			Aid       int    `json:"aid"`
			Videos    int    `json:"videos"`
			Tid       int    `json:"tid"`
			Tname     string `json:"tname"`
			Copyright int    `json:"copyright"`
			Pic       string `json:"pic"`
			Title     string `json:"title"`
			Pubdate   int    `json:"pubdate"`
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
				Mid  int    `json:"mid"`
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
			Bvid        string      `json:"bvid"`
			SeasonType  int         `json:"season_type"`
			IsOgv       bool        `json:"is_ogv"`
			OgvInfo     interface{} `json:"ogv_info"`
			RcmdReason  string      `json:"rcmd_reason"`
			EnableVt    int         `json:"enable_vt"`
			AiRcmd      struct {
				Id      int    `json:"id"`
				Goto    string `json:"goto"`
				Trackid string `json:"trackid"`
				UniqId  string `json:"uniq_id"`
			} `json:"ai_rcmd"`
			MissionId   int    `json:"mission_id,omitempty"`
			UpFromV2    int    `json:"up_from_v2,omitempty"`
			PubLocation string `json:"pub_location,omitempty"`
			SeasonId    int    `json:"season_id,omitempty"` // 视频所属合集的id
			FirstFrame  string `json:"first_frame,omitempty"`
			RedirectUrl string `json:"redirect_url,omitempty"`
		} `json:"Related"`
		Spec     interface{} `json:"Spec"`
		HotShare struct {
			Show bool          `json:"show"`
			List []interface{} `json:"list"`
		} `json:"hot_share"`
		Elec      interface{} `json:"elec"`
		Recommend interface{} `json:"recommend"`
		Emergency struct {
			NoLike  bool `json:"no_like"`
			NoCoin  bool `json:"no_coin"`
			NoFav   bool `json:"no_fav"`
			NoShare bool `json:"no_share"`
		} `json:"emergency"`
		ViewAddit struct {
			Field1 bool `json:"63"`
			Field2 bool `json:"64"`
			Field3 bool `json:"69"`
			Field4 bool `json:"71"`
			Field5 bool `json:"72"`
		} `json:"view_addit"`
		Guide     interface{} `json:"guide"`
		QueryTags interface{} `json:"query_tags"`
		IsOldUser bool        `json:"is_old_user"`
	} `json:"data"`
}

// VideoDetailResponse实现responseCheck接口
func (vd *VideoDetailResponse) getCode() int {
	return vd.Code
}
func (vd *VideoDetailResponse) bindJSON(data []byte) error {
	return json.Unmarshal(data, vd)
}
func (vd *VideoDetailResponse) BindJSON(data []byte) error {
	return vd.bindJSON(data)
}

type videoDetail struct {
}

func (receiver videoDetail) getRequest(byid string) *http.Request {
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/view/detail", nil)
	q := request.URL.Query()
	q.Add("bvid", byid)
	request.URL.RawQuery = q.Encode()
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edg/116.0.1938.69")
	//request.Header.Add("Cookie", bilibiliCookies.cookies)
	return request
}

func (receiver videoDetail) getResponse(bvid string) *VideoDetailResponse {
	biliCookiesManager.getUser(DefaultCookies).flushCookies()
	if !biliCookiesManager.getUser(DefaultCookies).cookiesFail {
		return nil
	}
	response, err := http.DefaultClient.Do(receiver.getRequest(bvid))
	if err != nil {
		utils.ErrorLog.Println(err.Error())
		return nil
	}
	result := new(VideoDetailResponse)
	err = responseCodeCheck(response, result)
	if err != nil {
		return nil
	}
	return result
}

func GetVideoDetailByByte(bvid string) ([]byte, string) {
	response, err := http.DefaultClient.Do(videoDetail{}.getRequest(bvid))
	if err != nil {
		utils.ErrorLog.Println(err.Error())
		return nil, response.Request.URL.String()
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, response.Request.URL.String()
}
