package bilibili

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"videoDynamicAcquisition/utils"
)

var (
	Spider             BiliSpider
	bilibiliCookies    cookies
	dynamicVideoObject dynamicVideo
	dynamicBaseUrl     = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed"
	followingsBseUrl   = "https://api.bilibili.com/x/relation/followings"
	//latestBaseline     = "" // 836201790065082504
	wbiSignObj = wbiSign{}
)

func init() {
	bilibiliCookies = cookies{}
	Spider = BiliSpider{}
	dynamicVideoObject = dynamicVideo{}
	bilibiliCookies.readFile()
	wbiSignObj.lastUpdateTime = time.Now()
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
		utils.ErrorLog.Println("cookies失效")
		bilibiliCookies.cookiesFail = false
		bilibiliCookies.flushCookies()
		if bilibiliCookies.cookiesFail {
			time.Sleep(time.Second * 10)
			return errors.New("cookies失效")
		} else {
			utils.ErrorLog.Println("cookies失效，请更新cookies文件2")
			return errors.New("cookies失效，请更新cookies文件")
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
