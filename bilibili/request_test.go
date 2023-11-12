package bilibili

import (
	"fmt"
	"testing"
	"videoDynamicAcquisition/utils"
)

func TestHistory(t *testing.T) {
	utils.InitLog("C:\\Code\\GO\\videoDynamicSpider")
	vd := videoDetail{}
	vd.getResponse("BV117411r7R1")
}

func TestFollowings(t *testing.T) {
	utils.InitLog("E:\\GoCode\\videoDynamicAcquisition")
	f := followings{
		pageNumber: 1,
	}
	fmt.Printf("%v\n", f.getResponse(0))
}
