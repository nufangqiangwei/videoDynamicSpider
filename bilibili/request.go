package bilibili

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
)

const (
	dynamicBaseUrl   = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed"
	followingsBseUrl = "https://api.bilibili.com/x/relation/followings"
	userAgent        = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
)

type responseCheck interface {
	getCode() int
	bindJSON([]byte) error
}

func responseCodeCheck(response *http.Response, apiResponseStruct responseCheck, user *cookies.UserCookie) error {
	defer response.Body.Close()
	requestUrl := response.Request.URL.String()
	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		log.ErrorLog.Println(requestUrl, "响应状态码错误", response.StatusCode)
		log.ErrorLog.Println(string(body))
		return errors.New(fmt.Sprintf("%s响应状态码错误:%d", requestUrl, response.StatusCode))
	}
	if err != nil {
		log.ErrorLog.Println(requestUrl, "读取响应失败")
		log.ErrorLog.Println(err.Error())
		return errors.New("读取响应失败")
	}
	err = apiResponseStruct.bindJSON(body)
	if err != nil {
		log.ErrorLog.Println(requestUrl, "json失败")
		log.ErrorLog.Println(err.Error())
		return errors.New("json失败")
	}
	code := apiResponseStruct.getCode()
	if code == -101 {
		log.ErrorLog.Printf("%s接口请求使用的%s用户cookies失效", requestUrl, user.GetUserName())
		// cookies失效
		user.InvalidCookies()
		return errors.New(fmt.Sprintf("%s:cookies失效，请更新cookies文件", user.GetUserName()))
	}
	if code == -352 {
		log.ErrorLog.Printf("%s接口请求使用的%s用户。352错误，拒绝访问", requestUrl, user.GetUserName())
		return errors.New("352错误，拒绝访问")
	}
	if code != 0 {
		log.ErrorLog.Printf("%s接口请求使用的%s用户。响应状态码错误%d", requestUrl, user.GetUserName(), code)
		log.ErrorLog.Println(string(body))
		return errors.New("响应状态码错误")
	}
	return nil
}
