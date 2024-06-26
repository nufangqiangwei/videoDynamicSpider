package bilibili

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/proxy"
)

type VideoInfoTypeEnum struct {
	DynamicTypeNone            string // 无效动态
	DynamicTypeForward         string // 动态转发
	DynamicTypeAv              string // 投稿视频
	DynamicTypePgc             string // 剧集（番剧、电影、纪录片）
	DynamicTypeCourses         string //
	DynamicTypeWord            string // 纯文字动态
	DynamicTypeDraw            string // 带图动态
	DynamicTypeArticle         string // 投稿专栏
	DynamicTypeMusic           string // 音乐
	DynamicTypeCommonSquare    string // 装扮,剧集点评,普通分享
	DynamicTypeCommonVertical  string //
	DynamicTypeLive            string // 直播间分享
	DynamicTypeMediaList       string // 收藏夹
	DynamicTypeCoursesSeason   string // 课程
	DynamicTypeCoursesBatch    string //
	DynamicTypeAd              string //
	DynamicTypeApplet          string //
	DynamicTypeSubscription    string //
	DynamicTypeLiveRcmd        string // 直播开播
	DynamicTypeBanner          string //
	DynamicTypeUgcSeason       string // 合集更新
	DynamicTypeSubscriptionNew string //
}

var DynamicInfoType = VideoInfoTypeEnum{
	DynamicTypeNone:            "DYNAMIC_TYPE_NONE",
	DynamicTypeForward:         "DYNAMIC_TYPE_FORWARD",
	DynamicTypeAv:              "DYNAMIC_TYPE_AV",
	DynamicTypePgc:             "DYNAMIC_TYPE_PGC",
	DynamicTypeCourses:         "DYNAMIC_TYPE_COURSES",
	DynamicTypeWord:            "DYNAMIC_TYPE_WORD",
	DynamicTypeDraw:            "DYNAMIC_TYPE_DRAW",
	DynamicTypeArticle:         "DYNAMIC_TYPE_ARTICLE",
	DynamicTypeMusic:           "DYNAMIC_TYPE_MUSIC",
	DynamicTypeCommonSquare:    "DYNAMIC_TYPE_COMMON_SQUARE",
	DynamicTypeCommonVertical:  "DYNAMIC_TYPE_COMMON_VERTICAL",
	DynamicTypeLive:            "DYNAMIC_TYPE_LIVE",
	DynamicTypeMediaList:       "DYNAMIC_TYPE_MEDIALIST",
	DynamicTypeCoursesSeason:   "DYNAMIC_TYPE_COURSES_SEASON",
	DynamicTypeCoursesBatch:    "DYNAMIC_TYPE_COURSES_BATCH",
	DynamicTypeAd:              "DYNAMIC_TYPE_AD",
	DynamicTypeApplet:          "DYNAMIC_TYPE_APPLET",
	DynamicTypeSubscription:    "DYNAMIC_TYPE_SUBSCRIPTION",
	DynamicTypeLiveRcmd:        "DYNAMIC_TYPE_LIVE_RCMD",
	DynamicTypeBanner:          "DYNAMIC_TYPE_BANNER",
	DynamicTypeUgcSeason:       "DYNAMIC_TYPE_UGC_SEASON",
	DynamicTypeSubscriptionNew: "DYNAMIC_TYPE_SUBSCRIPTION_NEW",
}

type updateVideoNumberResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		UpdateNum int `json:"update_num"`
	} `json:"data"`
}

type DynamicResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		HasMore        bool               `json:"has_more"` // 是否有更多数据
		Items          []DynamicVideoInfo `json:"items"`
		Offset         string             `json:"offset"`          // 偏移量 等于items中最后一条记录的id 获取下一页时使用
		UpdateBaseline string             `json:"update_baseline"` // 更新基线	等于items中第一条记录的id
		UpdateNum      int                `json:"update_num"`      // 本次获取获取到了多少条新动态,在更新基线以上的动态条数
	} `json:"data"`
}

type DynamicVideoInfo struct {
	Basic struct {
		CommentIdStr string `json:"comment_id_str"`
		CommentType  int    `json:"comment_type"`
		LikeIcon     struct {
			ActionUrl string `json:"action_url"`
			EndUrl    string `json:"end_url"`
			Id        int    `json:"id"`
			StartUrl  string `json:"start_url"`
		} `json:"like_icon"`
		RidStr string `json:"rid_str"`
	} `json:"basic"`
	IdStr   interface{} `json:"id_str"`
	Modules struct {
		ModuleAuthor struct {
			Avatar struct {
				ContainerSize struct {
					Height float64 `json:"height"`
					Width  float64 `json:"width"`
				} `json:"container_size"`
				FallbackLayers struct {
					IsCriticalGroup bool `json:"is_critical_group"`
					Layers          []struct {
						GeneralSpec struct {
							PosSpec struct {
								AxisX         float64 `json:"axis_x"`
								AxisY         float64 `json:"axis_y"`
								CoordinatePos int     `json:"coordinate_pos"`
							} `json:"pos_spec"`
							RenderSpec struct {
								Opacity int `json:"opacity"`
							} `json:"render_spec"`
							SizeSpec struct {
								Height float64 `json:"height"`
								Width  float64 `json:"width"`
							} `json:"size_spec"`
						} `json:"general_spec"`
						LayerConfig struct {
							IsCritical bool `json:"is_critical"`
							Tags       struct {
								AVATARLAYER struct {
								} `json:"AVATAR_LAYER"`
								GENERALCFG struct {
									ConfigType    int `json:"config_type"`
									GeneralConfig struct {
										WebCssStyle struct {
											BorderRadius string `json:"borderRadius"`
										} `json:"web_css_style"`
									} `json:"general_config"`
								} `json:"GENERAL_CFG"`
							} `json:"tags"`
						} `json:"layer_config"`
						Resource struct {
							ResImage struct {
								ImageSrc struct {
									Placeholder int `json:"placeholder"`
									Remote      struct {
										BfsStyle string `json:"bfs_style"`
										Url      string `json:"url"`
									} `json:"remote"`
									SrcType int `json:"src_type"`
								} `json:"image_src"`
							} `json:"res_image"`
							ResType int `json:"res_type"`
						} `json:"resource"`
						Visible bool `json:"visible"`
					} `json:"layers"`
				} `json:"fallback_layers"`
				Mid string `json:"mid"`
			} `json:"avatar"`
			Face           string `json:"face"`
			FaceNft        bool   `json:"face_nft"`
			Following      bool   `json:"following"`
			JumpUrl        string `json:"jump_url"`
			Label          string `json:"label"`
			Mid            int    `json:"mid"`
			Name           string `json:"name"`
			OfficialVerify struct {
				Desc string `json:"desc"`
				Type int    `json:"type"`
			} `json:"official_verify"`
			Pendant struct {
				Expire            int    `json:"expire"`
				Image             string `json:"image"`
				ImageEnhance      string `json:"image_enhance"`
				ImageEnhanceFrame string `json:"image_enhance_frame"`
				Name              string `json:"name"`
				Pid               int    `json:"pid"`
			} `json:"pendant"`
			PubAction       string `json:"pub_action"`
			PubLocationText string `json:"pub_location_text"`
			PubTime         string `json:"pub_time"`
			PubTs           int64  `json:"pub_ts"`
			Type            string `json:"type"`
			Vip             struct {
				AvatarSubscript    int    `json:"avatar_subscript"`
				AvatarSubscriptUrl string `json:"avatar_subscript_url"`
				DueDate            int64  `json:"due_date"`
				Label              struct {
					BgColor               string `json:"bg_color"`
					BgStyle               int    `json:"bg_style"`
					BorderColor           string `json:"border_color"`
					ImgLabelUriHans       string `json:"img_label_uri_hans"`
					ImgLabelUriHansStatic string `json:"img_label_uri_hans_static"`
					ImgLabelUriHant       string `json:"img_label_uri_hant"`
					ImgLabelUriHantStatic string `json:"img_label_uri_hant_static"`
					LabelTheme            string `json:"label_theme"`
					Path                  string `json:"lofFilePath"`
					Text                  string `json:"text"`
					TextColor             string `json:"text_color"`
					UseImgLabel           bool   `json:"use_img_label"`
				} `json:"label"`
				NicknameColor string `json:"nickname_color"`
				Status        int    `json:"status"`
				ThemeType     int    `json:"theme_type"`
				Type          int    `json:"type"`
			} `json:"vip"`
		} `json:"module_author"`
		ModuleDynamic struct {
			Additional interface{} `json:"additional"`
			Desc       interface{} `json:"desc"`
			Major      struct {
				Archive struct {
					Aid   string `json:"aid"`
					Badge struct {
						BgColor string      `json:"bg_color"`
						Color   string      `json:"color"`
						IconUrl interface{} `json:"icon_url"`
						Text    string      `json:"text"`
					} `json:"badge"`
					Bvid           string `json:"bvid"`
					Cover          string `json:"cover"`
					Desc           string `json:"desc"`
					DisablePreview int    `json:"disable_preview"`
					DurationText   string `json:"duration_text"`
					JumpUrl        string `json:"jump_url"`
					Stat           struct {
						Danmaku string `json:"danmaku"`
						Play    string `json:"play"`
					} `json:"stat"`
					Title string `json:"title"`
					Type  int    `json:"type"`
				} `json:"archive"`
				Type string `json:"type"`
			} `json:"major"`
			Topic interface{} `json:"topic"`
		} `json:"module_dynamic"`
		ModuleMore struct {
			ThreePointItems []struct {
				Label string `json:"label"`
				Type  string `json:"type"`
			} `json:"three_point_items"`
		} `json:"module_more"`
		ModuleStat struct {
			Comment struct {
				Count     int  `json:"count"`
				Forbidden bool `json:"forbidden"`
			} `json:"comment"`
			Forward struct {
				Count     int  `json:"count"`
				Forbidden bool `json:"forbidden"`
			} `json:"forward"`
			Like struct {
				Count     int  `json:"count"`
				Forbidden bool `json:"forbidden"`
				Status    bool `json:"status"`
			} `json:"like"`
		} `json:"module_stat"`
	} `json:"modules"`
	Type    string `json:"type"`
	Visible bool   `json:"visible"`
}

type dynamicVideo struct {
	userCookie    *cookies.UserCookie
	requestNumber int
}

// getRequest 设置请求头
// mid 用户的id,不传入的时候获取自己的关注动态
// https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset=&host_mid=591402619&timezone_offset=-480&features=itemOpusStyle,listOnlyfans
// https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all?offset=836201790065082504&timezone_offset=-480&type=video&features=itemOpusStyle,listOnlyfans
func (b *dynamicVideo) getRequest(mid int, offset string) *http.Request {
	url := ""
	if mid == 0 {
		url = dynamicBaseUrl + "/all"
	} else {
		url = dynamicBaseUrl + "/space"
	}
	request, _ := http.NewRequest("GET", url, nil)
	q := request.URL.Query()
	if offset != "" {
		q.Add("offset", offset)
	}
	if mid == 0 {
		q.Add("type", "video")
	} else {
		q.Add("host_mid", strconv.Itoa(mid))
	}

	request.URL.RawQuery = q.Encode()
	if b.userCookie == nil {
		b.userCookie = getDefaultUser()
	}
	request.Header.Add("Cookie", b.userCookie.GetCookies())
	return request
}

func (b *dynamicVideo) getUpdateVideoNumber(updateBaseline string) int {
	log.Info.Println("获取更新视频数量")
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all/update?type=video&update_baseline="+updateBaseline, nil)
	request.Header.Add("Cookie", b.userCookie.GetCookies())
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.ErrorLog.Println(err.Error())
		return 0
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		log.ErrorLog.Println("响应状态码错误", response.StatusCode)
		log.ErrorLog.Println(string(body))
		return 0
	}
	updateResponse := new(updateVideoNumberResponse)
	err = json.Unmarshal(body, updateResponse)
	if err != nil {
		log.ErrorLog.Println(err.Error())
		return 0
	}
	return updateResponse.Data.UpdateNum
}

func (b *dynamicVideo) getResponse(retriesNumber int, mid int, offset string, useProxy bool) (dynamicResponseBody *DynamicResponse) {
	if retriesNumber > 3 {
		return dynamicResponseBody
	}
	b.userCookie.FlushCookies()
	if !b.userCookie.GetStatus() {
		return dynamicResponseBody
	}

	response, err := proxy.GetClient(useProxy).Do(b.getRequest(mid, offset))
	if err != nil {
		log.ErrorLog.Println(err.Error())
		if useProxy && b.requestNumber < 1 {
			b.requestNumber++
			return b.getResponse(retriesNumber, mid, offset, useProxy)
		}
		return
	}
	dynamicResponseBody = &DynamicResponse{}
	err = responseCodeCheck(response, dynamicResponseBody, b.userCookie)
	if err != nil {
		if useProxy && b.requestNumber > 1 {
			b.requestNumber++
			return b.getResponse(retriesNumber, mid, offset, useProxy)
		}
		return nil
	}
	return dynamicResponseBody
}
func saveDynamicResponse(data []byte, mid int, offset string) {
	os.Mkdir(fmt.Sprintf("%s\\%d", baseStruct.RootPath, mid), os.ModePerm)
	fileName := fmt.Sprintf("%s\\%d\\bilibili-%s.json", baseStruct.RootPath, mid, offset)
	err := os.WriteFile(fileName, data, 0666)
	if err != nil {
		log.ErrorLog.Println("写文件失败")
		log.ErrorLog.Println(err.Error())
	}
}

// updateVideoNumberResponse实现responseCheck接口
func (u *updateVideoNumberResponse) getCode() int {
	return u.Code
}
func (u *updateVideoNumberResponse) bingJSON(body []byte) error {
	return json.Unmarshal(body, u)
}

// dynamicResponse实现responseCheck接口
func (d *DynamicResponse) getCode() int {
	return d.Code
}
func (d *DynamicResponse) bindJSON(body []byte) error {
	return json.Unmarshal(body, d)
}

func GetAuthorDynamic(uid string, offset string, user *cookies.UserCookie) *DynamicResponse {
	if user == nil {
		user = getDefaultUser()
	}
	b := dynamicVideo{
		userCookie: user,
	}
	auid, err := strconv.Atoi(uid)
	if err != nil {
		println(err.Error())
		return nil
	}
	return b.getResponse(0, auid, offset, true)
}
