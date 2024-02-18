package main

import (
	"github.com/gin-gonic/gin"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

func getAuthorFollowList(gtx *gin.Context) {
	webName := gtx.Query("webName")
	if webName == "" {
		gtx.JSONP(400, map[string]string{"msg": "webName不能为空"})
		gtx.Abort()
		return
	}
	var result []models.Author
	models.GormDB.Model(&models.Author{}).Joins("inner join web_site on author.web_site_id = web_site.id").Where("web_site.web_name = ? and (author.follow = 1 or author.crawl = 1)", webName).Find(&result)
	if result == nil {
		result = make([]models.Author, 0)
	}
	gtx.JSONP(200, result)
}

type migrateFollowAuthorRequestBody struct {
	WebName   string `json:"webName"`
	AuthorUid string `json:"authorUid"`
}

// migrateFollowAuthor 将主号的关注移动到小号上获取动态，省的自己去爬了
func migrateFollowAuthor(gtx *gin.Context) {
	requestBody := migrateFollowAuthorRequestBody{}
	err := gtx.BindJSON(&requestBody)
	if err != nil {
		gtx.JSONP(400, map[string]string{"msg": "参数错误"})
		return
	}
	err = bilibili.RelationAuthor(bilibili.FollowAuthor, requestBody.AuthorUid, "xiaohao")
	if err != nil {
		gtx.JSONP(400, map[string]string{"msg": "关注失败"})
		utils.ErrorLog.Printf("子账号添加关注失败,%s", err.Error())
		return
	}
}
