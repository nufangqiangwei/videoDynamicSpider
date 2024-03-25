package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"path"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/proxy"
	"videoDynamicAcquisition/utils"
)

var (
	config *utils.Config
	ginLog log.LogInputFile
)

func readConfig() error {
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

func checkToken(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSONP(403, map[string]string{"msg": "token错误"})
		ctx.Abort()
		return
	}
	jwt := Check{}
	err := proxy.DecryptToken(token, config.AesKey, config.AesIv, &jwt)
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

func main() {
	err := readConfig()
	if err != nil {
		log.ErrorLog.Printf("读取初始化配置文件出错")
		os.Exit(404)
	}
	defaultApp := gin.New()
	defaultApp.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: ginLog.File}), ginRecoveryLogging(), func(context *gin.Context) {
		ginLog.WriterObject.Println("\n+++++++++++++++++++\n")
	})

	proxyPath := defaultApp.Group("", checkToken)
	proxyPath.POST(proxy.AuthorVideoList.Path, getAuthorAllVideo)
	proxyPath.POST(proxy.SyncVideoListDetail.Path, getVideoDetailApi)
	proxyPath.GET(proxy.GetTaskStatus.Path, getTaskStatus)
	proxyPath.GET(proxy.VideoDetail.Path, getOneVideoInfo)

	if err = defaultApp.Run(fmt.Sprintf(":%d", config.ProxyWebServerLocalPort)); err != nil {
		ginLog.WriterObject.Println("GIN 框架出错：", err.Error())
		return
	}
}

func ginRecoveryLogging() gin.HandlerFunc {
	return gin.RecoveryWithWriter(ginLog.File, func(c *gin.Context, err any) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		if e, ok := err.(error); ok {
			c.String(200, e.Error())
		}
		if e, ok := err.(string); ok {
			c.String(500, e)
		}
		c.Abort()
	})
}
