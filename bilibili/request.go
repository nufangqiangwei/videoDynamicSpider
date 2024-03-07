package bilibili

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/utils"
)

const (
	dynamicBaseUrl   = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed"
	followingsBseUrl = "https://api.bilibili.com/x/relation/followings"
)

type responseCheck interface {
	getCode() int
	bindJSON([]byte) error
}

func responseCodeCheck(response *http.Response, apiResponseStruct responseCheck, user cookies.UserCookie) error {
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
		utils.ErrorLog.Printf("%s:cookies失效", user.GetUserName())
		user.InvalidCookies()
		if user.GetStatus() {
			time.Sleep(time.Second * 10)
			return errors.New(fmt.Sprintf("%s:cookies失效", user.GetUserName()))
		} else {
			utils.ErrorLog.Printf("%s:cookies失效，请更新cookies文件2", user.GetUserName())
			return errors.New(fmt.Sprintf("%s:cookies失效，请更新cookies文件", user.GetUserName()))
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
