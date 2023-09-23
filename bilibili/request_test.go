package bilibili

import (
	"fmt"
	"testing"
	"videoDynamicAcquisition/utils"
)

func TestHistory(t *testing.T) {
	utils.InitLog("E:\\GoCode\\videoDynamicAcquisition")
	a := Spider.GetVideoList("") // 844412517694308370
	fmt.Printf("%+v", a)
}

func TestFollowings(t *testing.T) {
	utils.InitLog("E:\\GoCode\\videoDynamicAcquisition")
	f := followings{
		pageNumber: 1,
	}
	fmt.Printf("%v\n", f.getResponse(0))
}
