package bilibili

import (
	"testing"
	"videoDynamicAcquisition/utils"
)

func TestHistory(t *testing.T) {
	utils.InitLog("E:\\GoCode\\videoDynamicAcquisition")
	a := historyRequest{}
	data := a.getResponse(0, 0)
	if data == nil {
		println("获取历史记录失败")
		return
	}
	data = a.getResponse(data.Data.Cursor.Max, data.Data.Cursor.ViewAt)
	if data == nil {
		println("获取历史记录失败")
		return
	}
	for _, info := range data.Data.List {
		println(info.Title)
	}
}
