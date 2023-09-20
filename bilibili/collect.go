package bilibili

// https://github.com/SocialSisterYi/bilibili-API-collect/blob/ffa25ba78dc8f4ed8624f11e3b6f404cb799674f/docs/fav/list.md api文档
// https://api.bilibili.com/x/v3/fav/folder/created/list-all?up_mid=10932398 获取收藏夹列表
// https://api.bilibili.com/x/v3/fav/folder/collected/list?pn=1&ps=20&up_mid=10932398&platform=web 获取收藏和订阅列表

type getCollectListResponse struct {
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

type getSubscriptionListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Count int `json:"count"`
		List  []struct {
			Id    int    `json:"id"`
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

type collectAllVideoListResponse struct {
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
			MediaCount int    `json:"media_count"`
		} `json:"info"`
		Medias []struct {
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
			Ctime   int         `json:"ctime"`
			Pubtime int         `json:"pubtime"`
			FavTime int         `json:"fav_time"`
			BvId    string      `json:"bv_id"`
			Bvid    string      `json:"bvid"`
			Season  interface{} `json:"season"`
		} `json:"medias"`
		HasMore bool `json:"has_more"`
	} `json:"data"`
}
