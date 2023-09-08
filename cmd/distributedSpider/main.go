package main

import (
	"github.com/gin-gonic/gin"
	"path"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

type requestBody struct {
	WebSite string `json:"webSite"`
	API     string `json:"api"`
}

func getWebSiteData(ctx *gin.Context) {

}

func main() {
	utils.InitLog(baseStruct.RootPath)
	models.InitDB(path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))

	server := gin.Default()
	server.POST("/getWebSiteData", getWebSiteData)
	server.Run("localhost:8001")
}
