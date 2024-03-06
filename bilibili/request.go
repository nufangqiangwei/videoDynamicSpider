package bilibili

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"videoDynamicAcquisition/utils"
)

const (
	dynamicBaseUrl   = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed"
	followingsBseUrl = "https://api.bilibili.com/x/relation/followings"
)

var (
	Spider             BiliSpider
	biliCookiesManager cookiesManager
	dynamicVideoObject dynamicVideo

	//latestBaseline     = "" // 836201790065082504
	wbiSignObj = wbiSign{}
)

func init() {
	Spider = BiliSpider{}
	biliCookiesManager = cookiesManager{
		cookiesMap: make(map[string]*cookies),
	}
	biliCookiesManager.flushCookies()
	wbiSignObj.lastUpdateTime = time.Now()
	for _, c := range biliCookiesManager.cookiesMap {
		dynamicVideoObject = dynamicVideo{
			userCookie: *c,
		}
		break
	}
	if dynamicVideoObject.userCookie.cookies == "" {
		panic("缺少cookies文件")
	}

}

type responseCheck interface {
	getCode() int
	bindJSON([]byte) error
}

func responseCodeCheck(response *http.Response, apiResponseStruct responseCheck) error {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		utils.ErrorLog.Println("响应状态码错误", response.StatusCode)
		utils.ErrorLog.Println(string(body))
		return errors.New(fmt.Sprintf("响应状态码错误:%d", response.StatusCode))
	}
	if err != nil {
		utils.ErrorLog.Println("读取响应失败")
		utils.ErrorLog.Println(err.Error())
		return errors.New("读取响应失败")
	}
	err = apiResponseStruct.bindJSON(body)
	if err != nil {
		utils.ErrorLog.Println("json失败")
		utils.ErrorLog.Println(err.Error())
		return errors.New("json失败")
	}
	code := apiResponseStruct.getCode()
	if code == -101 {
		// cookies失效

		requestCookies := response.Request.Cookies()
		buvid4 := ""
		for _, c := range requestCookies {
			if c.Name == "buvid4" {
				buvid4 = c.Value
			}
		}
		user := biliCookiesManager.cookiesGetUserName("buvid4", buvid4)
		utils.ErrorLog.Printf("%s:cookies失效", user)
		biliCookiesManager.getUser(user).cookiesFail = false
		biliCookiesManager.flushCookies()
		if biliCookiesManager.getUser(user).cookiesFail {
			time.Sleep(time.Second * 10)
			return errors.New(fmt.Sprintf("%s:cookies失效", user))
		} else {
			utils.ErrorLog.Printf("%s:cookies失效，请更新cookies文件2", user)
			return errors.New(fmt.Sprintf("%s:cookies失效，请更新cookies文件", user))
		}
	}
	if code == -352 {
		utils.ErrorLog.Println("352错误，拒绝访问")
		return errors.New("352错误，拒绝访问")
	}
	if code != 0 {
		utils.ErrorLog.Println("响应状态码错误", code)
		utils.ErrorLog.Println(string(body))
		return errors.New("响应状态码错误")
	}
	return nil
}
