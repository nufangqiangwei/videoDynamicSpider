package bilibili

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
)

const (
	dynamicBaseUrl   = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed"
	followingsBseUrl = "https://api.bilibili.com/x/relation/followings"
)

type responseCheck interface {
	getCode() int
	bindJSON([]byte) error
}

func responseCodeCheck(response *http.Response, apiResponseStruct responseCheck, user *cookies.UserCookie) error {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		log.ErrorLog.Println("响应状态码错误", response.StatusCode)
		log.ErrorLog.Println(string(body))
		return errors.New(fmt.Sprintf("响应状态码错误:%d", response.StatusCode))
	}
	if err != nil {
		log.ErrorLog.Println("读取响应失败")
		log.ErrorLog.Println(err.Error())
		return errors.New("读取响应失败")
	}
	err = apiResponseStruct.bindJSON(body)
	if err != nil {
		log.ErrorLog.Println("json失败")
		log.ErrorLog.Println(err.Error())
		return errors.New("json失败")
	}
	code := apiResponseStruct.getCode()
	if code == -101 {
		// cookies失效
		log.ErrorLog.Printf("%s:cookies失效", user.GetUserName())
		user.InvalidCookies()
		if user.GetStatus() {
			time.Sleep(time.Second * 10)
			return errors.New(fmt.Sprintf("%s:cookies失效", user.GetUserName()))
		} else {
			log.ErrorLog.Printf("%s:cookies失效，请更新cookies文件2", user.GetUserName())
			return errors.New(fmt.Sprintf("%s:cookies失效，请更新cookies文件", user.GetUserName()))
		}
	}
	if code == -352 {
		log.ErrorLog.Printf("%s用户。352错误，拒绝访问", user.GetUserName())
		return errors.New("352错误，拒绝访问")
	}
	if code != 0 {
		log.ErrorLog.Printf("%s用户。响应状态码错误%d", user.GetUserName(), code)
		log.ErrorLog.Println(string(body))
		return errors.New("响应状态码错误")
	}
	return nil
}
