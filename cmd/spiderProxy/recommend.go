package main

import (
	"github.com/gin-gonic/gin"
	"videoDynamicAcquisition/models"
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

func indexVideoRecommend(gtx *gin.Context) {
	models.GormDB.Exec("select * from")
}
