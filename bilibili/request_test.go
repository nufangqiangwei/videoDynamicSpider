package bilibili

import (
	"encoding/json"
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
func TestDynamic(t *testing.T) {
	utils.InitLog("E:\\GoCode\\videoDynamicAcquisition\\bilibili")
	dynamicVideoObject = dynamicVideo{}
	dynamicVideoObject.getResponse(0, 0, "")
}

/*
{"code":-101,"message":"账号未登录","ttl":1}
*/
func TestJSONDynamic(t *testing.T) {
	body := []byte(`{"code":-101,"message":"账号未登录","ttl":1}`)
	a := dynamicResponse{}
	err := json.Unmarshal(body, &a)
	if err != nil {
		print(err.Error())
		return
	}
	err = a.bindJSON(body)
	if err != nil {
		print(err.Error())
		return
	}
	fmt.Printf("%v\n", a)
}
