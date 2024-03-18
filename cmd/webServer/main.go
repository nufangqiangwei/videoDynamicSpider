package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"path"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

const (
	ExpirationTime   = 3600 * 24
	flushCookiesTime = 1800
)

type Check struct {
	Token          string `json:"token"`
	UserId         int64  `json:"userId"`
	ExpirationTime int64  `json:"expirationTime"`
	Ip             string `json:"ip"`
	LoginTime      int64  `json:"loginTime"`
}

func (c Check) String() string {
	data := make([]byte, 0)
	data, err := json.Marshal(&c)
	if err != nil {
		log.ErrorLog.Printf("Check模块json失败:%s", err.Error())
	}
	return string(data)
}

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
		ctx.Abort()
		return
	}
	jwt := Check{}
	err := utils.DecryptToken(token, config.AesKey, config.AesIv, &jwt)
	if err != nil {
		log.ErrorLog.Printf("解压token失败：%s", err.Error())
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		ctx.Abort()
		return
	}
	if jwt.Token != config.Token {
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		ctx.Abort()
		return
	}
	timeNow := time.Now().Unix()
	if timeNow-jwt.ExpirationTime > 60 {
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		ctx.Abort()
		return
	}
	ctx.Next()
}

func checkUser(ctx *gin.Context) {
	cookies, err := ctx.Cookie("info")
	if err != nil {
		ctx.JSONP(403, logoutResponse)
		ctx.Abort()
		return
	}
	userCookies := Check{}
	err = utils.DecryptToken(cookies, config.AesKey, config.AesIv, &userCookies)
	if err != nil {
		log.ErrorLog.Printf("解压用户cookies失败:%s", err.Error())
		ctx.JSONP(411, logoutResponse)
		ctx.Abort()
		return
	}
	//loginTime, err := models.GetUserLastLoginTime(userCookies.UserId)
	//if err != nil {
	//	utils.ErrorLog.Printf("redis获取失败:%s", err.Error())
	//	ctx.JSONP(403, map[string]string{"msg": "token错误2"})
	//	ctx.Abort()
	//	return
	//}
	//if loginTime != userCookies.ExpirationTime {
	//	ctx.JSONP(403, map[string]string{"msg": "非法cookies"})
	//	ctx.Abort()
	//	return
	//}
	if (time.Now().Unix() - userCookies.LoginTime) > ExpirationTime {
		ctx.JSONP(411, logoutResponse)
		ctx.Abort()
		return
	}

	user := models.User{}
	models.GormDB.Model(&models.User{}).Where("id = ?", userCookies.UserId).First(&user)

	// cookies即将过期刷新cookies
	if (ExpirationTime - (time.Now().Unix() - userCookies.ExpirationTime)) < flushCookiesTime {
		flushCookies := generateUserCookies(user)
		if flushCookies == "" {
			ctx.JSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		ctx.Header("Set-Cookie", flushCookies)
	}
	ctx.Set("user", user)
	ctx.Next()

}

func readConfig() error {
	//baseStruct.RootPath = "E:\\GoCode\\videoDynamicAcquisition\\cmd\\webServer"
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
	collection []models.VideoCollection
}

func main() {
	spiderManager = SpiderManager{
		collection: []models.VideoCollection{
			bilibili.Spider,
		},
	}
	err := readConfig()
	if err != nil {
		println(err.Error())
		os.Exit(4)
	}
	if len(config.AesKey) != 16 || len(config.AesIv) != 16 {
		println("aes key or iv error")
		os.Exit(4)
	}
	logBlockList := log.InitLog(baseStruct.RootPath, "database")
	var databaseLog log.LogInputFile
	for _, logBlock := range logBlockList {
		if logBlock.FileName == "database" {
			databaseLog = logBlock
			break
		}
	}
	if config.DB.HOST != "" {
		models.InitDB(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DB.User, config.DB.Password, config.DB.HOST, config.DB.Port, config.DB.DatabaseName), false, databaseLog.WriterObject)
		models.OpenRedis()
	}
	//go deleteFile()
	server := gin.Default()
	server.POST("register", registerUser)
	server.POST("login", userLogin)
	server.GET("supportWebSite", supportWebSiteLst)

	userApi := server.Group("user", checkUser)
	userApi.POST("resetPassword", resetUserPassword)
	userApi.POST("recommendVideo", bilibiliRecommendVideoSave)
	userApi.POST("uploadStaticFile", deployWebSIteHtmlFile)
	userApi.GET("manageCookies", getUserManageCookiesWebSite)
	userApi.GET("manageAccount", getUserFollowAuthor)
	userApi.POST("uploadAccountCookies", uploadWebCookies)

	proxyPath := server.Group("proxy", checkToken)
	proxyPath.POST(baseStruct.AuthorVideoList, getAuthorAllVideo)
	proxyPath.POST(baseStruct.VideoDetail, getVideoDetailApi)
	proxyPath.GET("getTaskStatus", getTaskStatus)

	video := server.Group("video", checkUser)
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

var successResponse = BaseResponse{
	Msg: "ok",
}
var logoutResponse = BaseResponse{
	Msg:  "logout",
	Code: 403,
}

/*
{"msg":"ok","code":0,"data":null}
*/
