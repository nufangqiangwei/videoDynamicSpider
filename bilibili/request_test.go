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
	models.InitDB("spider:p0o9i8u7@tcp(database:3306)/video?charset=utf8mb4&parseTime=True&loc=Local", false, nil)
	//cookies.DataSource = models.WebSiteCookies{}
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
	a := DynamicResponse{}
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

func TestBiliSpider_GetUserDynamic(t *testing.T) {
	cookies.FlushAllCookies()
	dynamicBaseLineMap = map[string]int64{
		"卢生啊": 1711268019, "干煸花椰菜": 1711322236,
	}
	historyBaseLineMap = map[string]int64{
		"卢生啊": 1709186160, "干煸花椰菜": 1711322124,
	}
	resultChan := make(chan models.Video)
	closeChan := make(chan models.TaskClose)
	go Spider.GetVideoList(resultChan, closeChan)
	for {
		select {
		case v := <-resultChan:
			fmt.Printf("标题:%s\n", v.Title)
		case c := <-closeChan:
			fmt.Printf("关闭:%v\n", c)
			return
		}
	}
}
