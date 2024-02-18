package bilibili

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/utils"
)

func TestMain(m *testing.M) {
	baseStruct.RootPath = "E:\\GoCode\\videoDynamicAcquisition"
	utils.InitLog(baseStruct.RootPath)
	biliCookiesManager.flushCookies()
	os.Exit(m.Run())
}

func TestHistory(t *testing.T) {
	vd := videoDetail{}
	vd.getResponse("BV117411r7R1")
}

func TestFollowings(t *testing.T) {
	f := followings{
		pageNumber: 1,
	}
	fmt.Printf("%v\n", f.getResponse(0))
}
func TestDynamic(t *testing.T) {
	dynamicVideoObject = dynamicVideo{}
	response := dynamicVideoObject.getResponse(0, 0, "", biliCookiesManager.getUser(DefaultCookies))
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

func TestGetFollowList(t *testing.T) {
	var (
		total     = 0
		maxPage   = 1
		f         followings
		localFile utils.WriteFile
	)
	localFile = utils.WriteFile{
		FolderPrefix: []string{baseStruct.RootPath},
		FileName: func(s string) string {
			return "followList.json"
		},
	}
	f = followings{}
	f.getFollowings(1)
	for {
		response := f.getResponse(0)
		if response == nil {
			response = &followingsResponse{}
		}
		if total == 0 {
			total = response.Data.Total
			if total%20 == 0 {
				maxPage = total / 20
			} else {
				maxPage = (total / 20) + 1
			}
		}
		x, _ := json.Marshal(response)
		localFile.WriteLine(x)
		if f.pageNumber >= maxPage {
			break
		}
		f.pageNumber++
		time.Sleep(time.Second * 3)
	}
}

func TestRelationAuthor(t *testing.T) {
	err := RelationAuthor(FollowAuthor, "3493118584293963", DefaultCookies)
	if err != nil {
		println(err.Error())
		return
	}
}
