package main

import (
	"github.com/gin-gonic/gin"
	"videoDynamicAcquisition/models"
)

func indexVideoRecommend(gtx *gin.Context) {
	models.GormDB.Raw(`select  * from author a
          inner join video_author va on a.id = va.author_id
          inner join  video v on va.video_id = v.id
          left join video_history vh on vh.video_id = v.id
where (follow=1 or crawl=1 ) and (vh.id is null or (v.duration - vh.view_time )>20)
order by upload_time desc`)
}

func videoHistoryList(gtx *gin.Context) {
	lastViewTime := gtx.Query("lastTime")
	if lastViewTime == "" {
		lastViewTime = "0"
	}
	querySql := `select w.web_name,
		       w.web_video_base_url,
		       w.web_author_base_url,
		       v.title,
		       v.video_desc,
		       v.cover_url,
		       a.author_name,
		       a.avatar,
		       a.author_web_uid,
		       vh.view_time
		from video_history vh
		         inner join video v on v.id = vh.video_id
		         inner join video_author va on va.video_id = vh.video_id
		         inner join author a on a.id = va.author_id
		         inner join web_site w on v.web_site_id = w.id
		where vh.view_time > FROM_UNIXTIME(?)
		order by vh.view_time desc;`
	var result []map[string]interface{}
	models.GormDB.Raw(querySql, lastViewTime).Scan(&result)
}

type videoUpdateListRequestBody struct {
	Page        int `form:"page"`
	Size        int `form:"size"`
	MinDuration int `form:"minDuration"` // 最小时长
	MaxDuration int `form:"maxDuration"` // 最大时长
}

func videoUpdateList(gtx *gin.Context) {
	requestBody := videoUpdateListRequestBody{}
	gtx.ShouldBindQuery(&requestBody)
	if requestBody.Page == 0 {
		requestBody.Page = 1
	}
	if requestBody.Size == 0 {
		requestBody.Size = 10
	}
	if requestBody.Size > 30 {
		requestBody.Size = 30
	}
	if requestBody.MinDuration < 30 {
		requestBody.MinDuration = 30
	}
	if requestBody.MaxDuration == 0 {
		requestBody.MaxDuration = 600
	}
	if requestBody.MinDuration > requestBody.MaxDuration {
		requestBody.MinDuration, requestBody.MaxDuration = requestBody.MaxDuration, requestBody.MinDuration
	} else if requestBody.MinDuration == requestBody.MaxDuration {
		requestBody.MinDuration = requestBody.MinDuration - 10
		requestBody.MaxDuration = requestBody.MaxDuration + 10
	}
	result := make([]map[string]interface{}, 0)
	models.GormDB.Raw(`select w.web_name as webSite, 
       v.title,
       v.uuid,
       CAST(UNIX_TIMESTAMP(v.upload_time) AS UNSIGNED) as uploadTime,
       v.video_desc as videoDesc,
       v.cover_url as coverUrl,
       v.duration,
       a.author_name as authorName,
       a.author_web_uid as authorWebUid,
       a.avatar
		from video v
		 inner join video_author va on va.video_id = v.id
		 inner join author a on va.author_id = a.id 
		 inner join web_site w on v.web_site_id = w.id
		inner join follow f on f.author_id=a.id
		where duration > ? and duration < ?
		order by  v.upload_time desc 
		limit ? offset ?`, requestBody.MinDuration, requestBody.MaxDuration,
		requestBody.Size, (requestBody.Page-1)*requestBody.Size).Scan(&result)
	// and v.upload_time >= CURDATE() - INTERVAL 30 DAY
	response := BaseResponse{
		Code: 0,
		Data: result,
	}
	gtx.JSONP(200, response)
}
