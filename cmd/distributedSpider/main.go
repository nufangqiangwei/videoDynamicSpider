package main

import (
	"github.com/gin-gonic/gin"
	"path"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

type GetAuthorVideoRequestBody struct {
	Author         string `json:"author"`
	StartPageIndex string `json:"StartPageIndex"`
	EndPageIndex   string `json:"endPageIndex"`
}

func GetAuthorVideoList(ctx *gin.Context) {
	requestBody := GetAuthorVideoRequestBody{}
	err := ctx.BindJSON(&requestBody)
	if err != nil {
		utils.ErrorLog.Println(err.Error())
		ctx.JSON(403, map[string]string{"data": "请求参数错误"})
		return
	}
	ctx.JSON(200, bilibili.Spider.GetVideoList())
}

func main() {
	utils.InitLog(baseStruct.RootPath)
	models.InitDB(path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))

	server := gin.Default()
	server.POST("/getAuthorVideoList", GetAuthorVideoList)
	err := server.Run("localhost:8001")
	if err != nil {
		utils.ErrorLog.Println(err.Error())
	}
}
