package main

import (
	"github.com/gin-gonic/gin"
	"videoDynamicAcquisition/baseStruct"
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

}

func main() {
	utils.InitLog(baseStruct.RootPath)
	baseStruct.InitDB()

	server := gin.Default()
	server.POST("/getAuthorVideoList", GetAuthorVideoList)
	err := server.Run("localhost:8001")
	if err != nil {
		utils.ErrorLog.Println(err.Error())
	}
}
