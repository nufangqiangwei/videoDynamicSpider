package main

import (
	"github.com/gin-gonic/gin"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
)

func getAuthorFollowList(gtx *gin.Context) {
	webName := gtx.Query("webName")
	if webName == "" {
		gtx.JSONP(400, map[string]string{"msg": "webName不能为空"})
		gtx.Abort()
		return
	}
	var result []models.Author
	models.GormDB.Model(&models.Author{}).Joins("inner join web_site on author.web_site_id = web_site.id").Where("web_site.web_name = ? and author.crawl = 1", webName).Find(&result)
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
	user := cookies.GetUser(bilibili.Spider.GetWebSiteName().WebName, "xiaohao")
	err = bilibili.RelationAuthor(bilibili.FollowAuthor, requestBody.AuthorUid, *user)
	if err != nil {
		gtx.JSONP(400, map[string]string{"msg": "关注失败"})
		log.ErrorLog.Printf("子账号添加关注失败,%s", err.Error())
		return
	}
}

type followAuthorRequestBody struct {
	WebName        string `json:"webName"`
	AuthorUid      string `json:"authorUid"`
	SourceUserName string `json:"sourceUserName"`
	TargetUserName string `json:"targetUserName"`
}

func followAuthor(gtx *gin.Context) {
	requestBody := followAuthorRequestBody{}
	err := gtx.BindJSON(&requestBody)
	if err != nil {
		gtx.JSONP(400, map[string]string{"msg": "参数错误"})
		return
	}
	var (
		sourceUser *cookies.UserCookie
		targetUser *cookies.UserCookie
	)
	for _, webSiteInfo := range spiderManager.collection {
		if webSiteInfo.GetWebSiteName().WebName == requestBody.WebName {
			sourceUser = cookies.GetUser(webSiteInfo.GetWebSiteName().WebName, requestBody.SourceUserName)
			sourceUser = cookies.GetUser(webSiteInfo.GetWebSiteName().WebName, requestBody.TargetUserName)
			break
		}
	}
	if sourceUser == nil || targetUser == nil {
		gtx.JSONP(400, map[string]string{"msg": "账号错误"})
		return
	}
	err = bilibili.RelationAuthor(bilibili.FollowAuthor, requestBody.AuthorUid, *targetUser)
	if err != nil {
		gtx.JSONP(400, map[string]string{"msg": "关注失败"})
		log.ErrorLog.Printf("子账号添加关注失败,%s", err.Error())
		return
	}
	err = bilibili.RelationAuthor(bilibili.UnFollowAuthor, requestBody.AuthorUid, *sourceUser)
	if err != nil {
		gtx.JSONP(400, map[string]string{"msg": "取关失败"})
		log.ErrorLog.Printf("主账号取关失败,%s", err.Error())
		return
	}

}
