package bilibili

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
)

func TestMain(m *testing.M) {
	baseStruct.RootPath = "E:\\GoCode\\videoDynamicAcquisition"
	log.InitLog(baseStruct.RootPath)
	cookies.DataSource = models.WebSiteCookies{}
	cookies.FlushAllCookies()
	os.Exit(m.Run())
}

func TestHistory(t *testing.T) {
	vd := videoDetail{}
	vd.getResponse("BV117411r7R1")
}

func TestDynamic(t *testing.T) {
	dynamicVideoObject := dynamicVideo{}
	response := dynamicVideoObject.getResponse(0, 0, "")
	if response == nil {
		println("获取失败")
	} else {
		fmt.Printf("%v\n", response)
	}
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

func TestRelationAuthor(t *testing.T) {
	err := RelationAuthor(FollowAuthor, "3493118584293963", cookies.UserCookie{})
	if err != nil {
		println(err.Error())
		return
	}
}
func TestBVAV(t *testing.T) {
	av := Bv2Av("BV1bV411S7Le")
	println(av)
	bv := Av2Bv(411857180)
	println(bv)
}
