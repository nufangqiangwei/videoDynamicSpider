package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"path"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
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

func checkToken(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		return
	}
	decryptToken := utils.DecryptToken(token, config.AesKey)
	if decryptToken["token"] != config.Token {
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		return
	}
	requestTime := decryptToken["time"]
	timeNow := time.Now().Unix()
	requestTimeInt, err := strconv.ParseInt(requestTime, 10, 64)
	if err != nil {
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		return
	}
	if timeNow-requestTimeInt > 60 {
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		return
	}
	ctx.Next()
}

func readConfig() error {
	//baseStruct.RootPath = "E:\\GoCode\\videoDynamicAcquisition\\cmd\\spiderProxy"
	fileData, err := os.ReadFile(path.Join(baseStruct.RootPath, "config.json"))
	if err != nil {
		println(err.Error())
		return err
	}
	config = &utils.Config{}
	err = json.Unmarshal(fileData, config)
	if err != nil {
		println(err.Error())
		return err
	}
	fmt.Printf("%v\n", *config)
	return nil
}

var (
	spiderManager = SpiderManager{}
)

type SpiderManager struct {
	collection []baseStruct.VideoCollection
}

func main() {
	spiderManager = SpiderManager{
		collection: []baseStruct.VideoCollection{
			bilibili.Spider,
		},
	}
	err := readConfig()
	if err != nil {
		println(err.Error())
		return
	}
	utils.InitLog(baseStruct.RootPath)
	if config.DB.HOST != "" {
		models.InitDB(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DB.User, config.DB.Password, config.DB.HOST, config.DB.Port, config.DB.DatabaseName), false)
	}
	//go deleteFile()
	server := gin.Default()
	server.POST("recommendVideo", bilibiliRecommendVideoSave)
	server.POST("uploadStaticFile", deployWebSIteHtmlFile)

	proxyPath := server.Group("proxy", checkToken)
	proxyPath.POST(baseStruct.AuthorVideoList, getAuthorAllVideo)
	proxyPath.POST(baseStruct.VideoDetail, getVideoDetailApi)
	proxyPath.GET("getTaskStatus", getTaskStatus)

	video := server.Group("video", checkDBInit)
	video.GET("getFollowList", getAuthorFollowList)
	video.GET("migrateFollowAuthor", migrateFollowAuthor)
	video.GET("list", videoUpdateList)

	server.Run(fmt.Sprintf(":%d", config.ProxyWebServerLocalPort))
}

type BaseResponse struct {
	Msg  string      `json:"msg"`
	Code int64       `json:"code"`
	Data interface{} `json:"data"`
}
