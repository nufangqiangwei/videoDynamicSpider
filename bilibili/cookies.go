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
}

func (c *cookies) flushCookies() {
	if !c.cookiesFail || c.cookies == "" || c.lastFlushCookiesTime.Add(time.Hour*24).Before(time.Now()) {
		// 如果cookies失效并且上次刷新时间超过24小时，重新刷新cookies
		c.lastFlushCookiesTime = time.Now()
		c.readFile()
		if !c.cookiesFail {
			// cookies刷新失败
			utils.ErrorLog.Println("cookies失效，请更新cookies文件1")
		}
	}
}

func (c *cookies) readFile() {
	// 读取文件中的cookies
	filePath := path.Join(baseStruct.RootPath, "bilibiliCookies")
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
