package bilibili

import (
	"os"
	"path"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/utils"
)

type cookies struct {
	cookies              string
	lastFlushCookiesTime time.Time
	cookiesFail          bool
	fileName             string
}

const (
	cookiesFileFolder = "biliCookies"
	DefaultCookies    = "default"
)

func (c *cookies) flushCookies() {
	if !c.cookiesFail || c.cookies == "" || c.lastFlushCookiesTime.Add(time.Hour*24).Before(time.Now()) {
		// 如果cookies失效并且上次刷新时间超过24小时，重新刷新cookies
		c.lastFlushCookiesTime = time.Now()
		c.readFile()
		if !c.cookiesFail {
			// cookies刷新失败
			if utils.ErrorLog != nil {
				utils.ErrorLog.Println("cookies失效，请更新cookies文件")
			} else {
				println("cookies失效，请更新cookies文件")
			}
		}
	}
}

func (c *cookies) readFile() {
	// 读取文件中的cookies
	filePath := path.Join(baseStruct.RootPath, cookiesFileFolder, c.fileName)
	println("bilibili Cookies地址：", filePath)
	f, err := os.ReadFile(filePath)

	if err != nil {
		c.cookies = ""
		c.cookiesFail = false
		return
	}
	cookies := strings.TrimSpace(string(f))
	if !c.cookiesFail {
		c.cookiesFail = c.cookies != cookies
	}
	c.cookies = cookies
}

func (c *cookies) getCookiesKeyValue(keyName string) string {
	c.flushCookies()
	cookies := strings.Split(c.cookies, ";")
	for _, cookie := range cookies {
		cookie = strings.TrimSpace(cookie)
		if strings.HasPrefix(cookie, keyName) {
			return strings.Split(cookie, "=")[1]
		}
	}
	return ""
}

type cookiesManager struct {
	cookiesMap map[string]*cookies
}

func defaultUserCookies() *cookies {
	return biliCookiesManager.cookiesMap["default"]
}
func (cm *cookiesManager) flushCookies() {
	// 遍历文件夹下的所有cookies文件，刷新cookies
	files, err := os.ReadDir(path.Join(baseStruct.RootPath, cookiesFileFolder))
	if err != nil {
		println("读取cookies文件夹失败")
		println(err.Error())
		if utils.ErrorLog != nil {
			utils.ErrorLog.Println("读取cookies文件夹失败")
		} else {
			println("读取cookies文件夹失败")
		}
		return
	}
	for _, file := range files {
		c, ok := biliCookiesManager.cookiesMap[file.Name()]
		if !ok {
			c = &cookies{fileName: file.Name()}
			biliCookiesManager.cookiesMap[file.Name()] = c
		}
		c.flushCookies()
	}
	_, ok := biliCookiesManager.cookiesMap["default"]
	if !ok {
		biliCookiesManager.cookiesMap["default"] = &cookies{}
	}
}
func (cm *cookiesManager) getUser(key string) *cookies {
	c, ok := cm.cookiesMap[key]
	if !ok {
		c = &cookies{fileName: key}
		cm.cookiesMap[key] = c
	}
	return c
}
