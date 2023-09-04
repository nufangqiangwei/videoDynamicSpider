package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	timeWheel "github.com/nufangqiangwei/timewheel"
	"net/http"
	"path"
	"strconv"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
)

var (
	videoCollection []VideoCollection
	wheel           *timeWheel.TimeWheel
	spider          *Spider
)

type VideoCollection interface {
	GetWebSiteName() models.WebSite
	GetVideoList() []baseStruct.VideoInfo
}

type Spider struct {
	interval int64
}

func (s *Spider) getVideoInfo() {
	println("getVideoInfo")
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	for _, v := range videoCollection {
		website := v.GetWebSiteName()
		website.GetOrCreate(db)
		for _, video := range v.GetVideoList() {
			//fmt.Printf("video: %v\n", video)
			author := models.Author{AuthorName: video.AuthorName, WebSiteId: website.Id, AuthorWebUid: video.AuthorUuid}
			author.GetOrCreate(db)
			videoModel := models.Video{
				WebSiteId:  website.Id,
				AuthorId:   author.Id,
				Title:      video.Title,
				Desc:       video.Desc,
				Duration:   video.Duration,
				Url:        video.Url,
				Uuid:       video.VideoUuid,
				CoverUrl:   video.CoverUrl,
				UploadTime: video.PushTime,
				BiliOffset: video.Baseline,
			}
			videoModel.Save(db)
		}
	}
}

func (s *Spider) run(interface{}) {
	for _, v := range videoCollection {
		videoList := v.GetVideoList()
		for _, video := range videoList {
			fmt.Printf("video: %v\n", video)
		}
	}
	wheel.AppendOnceFunc(s.run, nil, "VideoSpider", timeWheel.Crontab{ExpiredTime: s.interval})
}

func getVideoList(ctx *gin.Context) {
	db, err := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	if err != nil {
		println(err.Error())
		ctx.JSONP(http.StatusInternalServerError, map[string]string{"msg": "数据库打开失败"})
		return
	}
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
	result := map[string][]baseStruct.VideoInfo{"data": []baseStruct.VideoInfo{}}
	if err != nil {
		println(err.Error())
		ctx.JSONP(200, result)
		return
	}
	for rows.Next() {
		v := baseStruct.VideoInfo{}
		err = rows.Scan(&v.VideoUuid, &v.CoverUrl, &v.Title, &v.Desc, &v.Duration, &v.AuthorName, &v.AuthorUuid, &v.PushTime)
		if err != nil {
			print("scan error")
			println(err.Error())
		}
		result["data"] = append(result["data"], v)
	}
	ctx.JSON(200, result)
}

func updateCookies(ctx *gin.Context) {

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
	//InitLog("E:\\GoCode\\videoDynamicAcquisition")
	models.InitDB(path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	//videoCollection = []VideoCollection{
	//	bilibili.MakeBilibiliSpider(),
	//}
	//spider = &Spider{interval: 60 * 5}
	//wheel = timeWheel.NewTimeWheel(&timeWheel.WheelConfig{
	//	IsRun: true,
	//	Log:   timewheelLog,
	//})
	//wheel.AppendOnceFunc(spider.run, nil, "VideoSpider", timeWheel.Crontab{ExpiredTime: spider.interval})
	//spider.getVideoInfo()
	//server := gin.Default()
	//server.Use(Cors())
	//server.GET("/getVideoList", getVideoList)
	//server.GET("/updateCookies", updateCookies)
	//server.Run("localhost:8001")
}
