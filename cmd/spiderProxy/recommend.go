package main

import (
	"github.com/gin-gonic/gin"
	"videoDynamicAcquisition/models"
)

func checkDBInit(gtx *gin.Context) {
	if models.GormDB == nil {
		gtx.JSONP(500, map[string]string{"msg": "数据库连接失败"})
		gtx.Abort()
		return
	}
	gtx.Next()
	return
}

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
