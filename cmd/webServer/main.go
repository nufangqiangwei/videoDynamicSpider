package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

func getVideoList(ctx *gin.Context) {
	db := baseStruct.CanUserDb()
	defer db.Close()

	webSiteName := ctx.DefaultQuery("webSite", "bilibili")
	pageSTR := ctx.DefaultQuery("page", "1")
	sizeSTR := ctx.DefaultQuery("size", "30")
	page, err := strconv.ParseInt(pageSTR, 10, 32)
	if err != nil {
		page = 1
	}
	size, err := strconv.ParseInt(sizeSTR, 10, 32)
	if err != nil {
		size = 30
	}
	rows, err := db.Query("select v.uuid, v.cover_url,v.title, v.video_desc,v.duration, a.author_name, a.author_web_uid,v.upload_time from video v left join author a on v.author_id = a.id where v.web_site_id = (select id from website where web_name = ?) order by v.upload_time desc limit ?,?;", webSiteName, (page-1)*size, page*size)
	result := map[string][]baseStruct.VideoInfo{"data": {}}
	if err != nil {
		utils.ErrorLog.Println(err.Error())
		ctx.JSONP(200, result)
		return
	}
	for rows.Next() {
		v := baseStruct.VideoInfo{}
		err = rows.Scan(&v.VideoUuid, &v.CoverUrl, &v.Title, &v.Desc, &v.Duration, &v.AuthorName, &v.AuthorUuid, &v.PushTime)
		if err != nil {
			utils.ErrorLog.Println("scan error")
			utils.ErrorLog.Println(err.Error())
		}
		result["data"] = append(result["data"], v)
	}
	ctx.JSON(200, result)
}

func updateCookies(ctx *gin.Context) {

}

func addAuthor(ctx *gin.Context) {
	db := baseStruct.CanUserDb()
	defer db.Close()

	authorId := ctx.DefaultQuery("authorId", "1")
	authorName := ctx.DefaultQuery("AuthorName", "30")
	avatarUrl := ctx.DefaultQuery("Avatar", "30")
	desc := ctx.DefaultQuery("Desc", "30")
	author := models.Author{
		WebSiteId:    1,
		AuthorWebUid: authorId,
		AuthorName:   authorName,
		Avatar:       avatarUrl,
		Desc:         desc,
		Follow:       false,
	}
	author.GetOrCreate(db)
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func main() {
	server := gin.Default()
	server.Use(Cors())
	server.GET("/getVideoList", getVideoList)
	server.GET("/updateCookies", updateCookies)
	err := server.Run("localhost:8001")
	if err != nil {
		utils.ErrorLog.Println(err.Error())
	}
}
