package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"videoDynamicAcquisition/models"
)

type bilbilDynamicResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		HasMore        bool              `json:"has_more"`
		Items          []bilbilVideoInfo `json:"items"`
		Offset         string            `json:"offset"`
		UpdateBaseline string            `json:"update_baseline"`
		UpdateNum      int               `json:"update_num"`
	} `json:"data"`
}

type bilbilVideoInfo struct {
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
	IdStr   string `json:"id_str"`
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

type bilibiliDynamicVideoUrl struct {
	baseUrl     string
	pageIndex   int
	dynamicType string
	cookies     string
}

func (b *bilibiliDynamicVideoUrl) getNextUrl() string {
	// https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all?type=video&page=1
	url := b.baseUrl + "?type=" + b.dynamicType + "&page=" + strconv.Itoa(b.pageIndex)
	b.pageIndex++
	return url
}

func (b *bilibiliDynamicVideoUrl) getRequest() *http.Request {
	request, _ := http.NewRequest("GET", "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all?type=video&page=1", nil)
	if b.cookies == "" {
		b.flushCookies()
	}
	request.Header.Add("Cookie", b.cookies)
	return request

}
func (b *bilibiliDynamicVideoUrl) flushCookies() bool {
	// 读取文件中的cookies
	f, err := os.ReadFile("bilibiliCookies")
	if err != nil {
		b.cookies = ""
		return false
	}
	cookies := string(f)
	isFlush := b.cookies == cookies
	b.cookies = cookies
	return isFlush
}

type bilibiliSpider struct {
	url                  *bilibiliDynamicVideoUrl
	lastFlushCookiesTime time.Time
	cookiesFail          bool
}

func makeBilibiliSpider() bilibiliSpider {
	bilibili := bilibiliSpider{}
	bilibili.url = &bilibiliDynamicVideoUrl{
		baseUrl:     "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all",
		pageIndex:   1,
		dynamicType: "video",
	}
	bilibili.url.flushCookies()
	return bilibili
}

func (bilibili bilibiliSpider) getWebSiteName() models.WebSite {
	return models.WebSite{
		WebName:          "bilibili",
		WebHost:          "https://www.bilibili.com/",
		WebAuthorBaseUrl: "https://space.bilibili.com/",
		WebVideoBaseUrl:  "https://www.bilibili.com/",
	}
}

func (bilibili bilibiliSpider) getResponse() *bilbilDynamicResponse {
	if bilibili.cookiesFail && bilibili.lastFlushCookiesTime.Add(time.Hour*24).Before(time.Now()) {
		// 如果cookies失效并且上次刷新时间超过24小时，重新刷新cookies
		bilibili.lastFlushCookiesTime = time.Now()
		bilibili.cookiesFail = !bilibili.url.flushCookies()
		if bilibili.cookiesFail {
			// cookies刷新失败
			log.Println("cookies失效，请更新cookies文件")
			return nil
		}
	}
	client := &http.Client{}
	response, err := client.Do(bilibili.url.getRequest())
	if err != nil {
		log.Println(err)
		return nil
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("读取响应失败")
		return nil
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		log.Println("响应状态码错误", response.StatusCode)
		log.Println(string(responseBytes))
		return nil
	}

	dynamicResponse := &bilbilDynamicResponse{}
	err = json.Unmarshal(responseBytes, dynamicResponse)
	if err != nil {
		log.Println(err)
		return nil
	}
	if dynamicResponse.Code == -101 {
		// cookies失效
		log.Println("cookies失效")
		if bilibili.lastFlushCookiesTime.Add(time.Hour * 24).Before(time.Now()) {
			// 24小时内不刷新cookies
			log.Println("刷新cookies")
			bilibili.lastFlushCookiesTime = time.Now()
			bilibili.cookiesFail = !bilibili.url.flushCookies()
			if bilibili.cookiesFail {
				// cookies刷新失败
				log.Println("cookies失效，请更新cookies文件")
				return nil
			}
			return bilibili.getResponse()
		}
	}
	bilibili.cookiesFail = false
	return dynamicResponse
}

func (bilibili bilibiliSpider) getVideoList() []VideoInfo {
	response := bilibili.getResponse()
	if response == nil {
		return []VideoInfo{}
	}
	if response.Data.Items == nil {
		return []VideoInfo{}
	}
	result := make([]VideoInfo, len(response.Data.Items))
	for index, info := range response.Data.Items {
		result[index] = VideoInfo{
			WebSite:    "bilibili",
			Title:      info.Modules.ModuleDynamic.Major.Archive.Title,
			Uuid:       info.Modules.ModuleDynamic.Major.Archive.Bvid,
			Url:        info.Modules.ModuleDynamic.Major.Archive.JumpUrl,
			CoverUrl:   info.Modules.ModuleDynamic.Major.Archive.Cover,
			AuthorName: info.Modules.ModuleAuthor.Name,
			AuthorUrl:  info.Modules.ModuleAuthor.JumpUrl,
			PushTime:   time.Unix(info.Modules.ModuleAuthor.PubTs, 0),
		}
	}
	return result
}
