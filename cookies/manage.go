package cookies

import (
	"strconv"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/utils"
)

const (
	cookiesFileFolder = "baseCookies"
	blankUserName     = "空用户"
)

var DataSource baseStruct.CookiesFlush = privateReadLocalFile{}

func printLog(args ...string) {
	if utils.ErrorLog == nil {
		println(args)
	} else {
		utils.ErrorLog.Println(args)
	}
}

type UserCookie struct {
	cookies              string
	lastFlushCookiesTime time.Time
	cookiesFail          bool
	fileName             string
	webSiteName          string
	dbPrimaryKeyId       int64
}

func (c *UserCookie) SetDBPrimaryKeyId(id int64) {
	c.dbPrimaryKeyId = id
}
func (c *UserCookie) GetDBPrimaryKeyId() int64 {
	return c.dbPrimaryKeyId
}

func (c *UserCookie) GetStatus() bool {
	return c.cookiesFail
}
func (c *UserCookie) InvalidCookies() {
	c.cookiesFail = true
	c.FlushCookies()
}

func (c *UserCookie) GetUserName() string {
	return c.fileName
}

func (c *UserCookie) GetCookies() string {
	return c.cookies
}

func (c *UserCookie) FlushCookies() {
	if !c.cookiesFail || c.cookies == "" || c.lastFlushCookiesTime.Add(time.Hour*24).Before(time.Now()) {
		// 如果cookies失效并且上次刷新时间超过24小时，重新刷新cookies
		if !c.cookiesFail {
			utils.Info.Printf("%scookies标注为失效", c.fileName)
		} else if c.cookies == "" {
			utils.Info.Printf("%scookies未加载", c.fileName)
		} else {
			utils.Info.Printf("%scookies距离上次加载已过24小时。上次加载时间%s", c.fileName, c.lastFlushCookiesTime.Format("2006.01.02 15:04:05"))
		}
		c.lastFlushCookiesTime = time.Now()
		c.readFile()
		if !c.cookiesFail {
			// cookies刷新失败
			if utils.ErrorLog != nil {
				utils.ErrorLog.Printf("cookies失效，请更新%scookies文件", c.fileName)
			} else {
				println("cookies失效，请更新", c.fileName, "cookies文件")
			}
		}
	}
}

func (c *UserCookie) setCookies(cookiesContext string) {
	c.cookies = cookiesContext
}

func (c *UserCookie) saveCookies() {
	err := DataSource.UpdateUserCookies(c.webSiteName, c.fileName, c.cookies, strconv.FormatInt(c.dbPrimaryKeyId, 10))
	if err != nil {
		utils.ErrorLog.Printf("更新%scookies文件失败", c.fileName)
	}
}

func (c *UserCookie) readFile() {
	// 读取文件中的cookies
	if c.fileName == blankUserName {
		return
	}

	cookies := DataSource.GetUserCookies(c.webSiteName, c.fileName)
	if !c.cookiesFail {
		c.cookiesFail = c.cookies != cookies
	}
	c.cookies = cookies
}

func (c *UserCookie) GetCookiesKeyValue(keyName string) string {
	c.FlushCookies()
	cookies := strings.Split(c.cookies, ";")
	for _, cookie := range cookies {
		cookie = strings.TrimSpace(cookie)
		if strings.HasPrefix(cookie, keyName) {
			return strings.Split(cookie, "=")[1]
		}
	}
	return ""
}

func NewDefaultUserCookie(webSiteName string) UserCookie {
	return UserCookie{
		fileName:    blankUserName,
		webSiteName: webSiteName,
	}
}

func NewTemporaryUserCookie(webSiteName, cookiesText string) UserCookie {
	return UserCookie{
		cookies:     cookiesText,
		fileName:    blankUserName,
		webSiteName: webSiteName,
	}
}

type WebSiteCookiesManager struct {
	cookiesMap  map[string]*UserCookie
	webSiteName string
}

func (wcm *WebSiteCookiesManager) FlushCookies() {
	for _, userInfo := range DataSource.UserList(wcm.webSiteName) {
		c, ok := wcm.cookiesMap[userInfo.UserName]
		if !ok {
			c = &UserCookie{fileName: userInfo.UserName, webSiteName: wcm.webSiteName, cookies: userInfo.Content}
			wcm.cookiesMap[userInfo.UserName] = c
		}

	}
}

var rootCookiesMap map[string]WebSiteCookiesManager

func FlushAllCookies() {
	if rootCookiesMap == nil {
		rootCookiesMap = make(map[string]WebSiteCookiesManager)
	}
	for _, webSiteName := range DataSource.WebSiteList() {
		w, ok := rootCookiesMap[webSiteName]
		if !ok {
			w = WebSiteCookiesManager{
				cookiesMap:  make(map[string]*UserCookie),
				webSiteName: webSiteName,
			}
			rootCookiesMap[webSiteName] = w
		}
		w.FlushCookies()
	}
}

func GetUser(weiSiteName, userName string) *UserCookie {
	w, ok := rootCookiesMap[weiSiteName]
	if !ok {
		w = WebSiteCookiesManager{
			cookiesMap:  make(map[string]*UserCookie),
			webSiteName: weiSiteName,
		}
		rootCookiesMap[weiSiteName] = w
	}
	c, ok := w.cookiesMap[userName]
	if !ok {
		c = &UserCookie{fileName: userName}
		w.cookiesMap[userName] = c
	}
	return c
}

func CookiesGetUserName(weiSiteName, cookieKey, cookieValue string) string {
	websiteData, ok := rootCookiesMap[weiSiteName]
	if !ok {
		return ""
	}
	for userName, cookiesObj := range websiteData.cookiesMap {
		if cookiesObj.GetCookiesKeyValue(cookieKey) == cookieValue {
			return userName
		}
	}
	return ""
}

func GetWebSiteUser(weiSiteName string) map[string]*UserCookie {
	websiteData, ok := rootCookiesMap[weiSiteName]
	if !ok {
		return map[string]*UserCookie{}
	}
	return websiteData.cookiesMap
}

func RangeCookiesMap(f func(weiSiteName, userName string, cookies *UserCookie)) {
	for weiSiteName, websiteData := range rootCookiesMap {
		for userName, cookies := range websiteData.cookiesMap {
			f(weiSiteName, userName, cookies)
		}
	}
}

func UpdateUserCookies(weiSiteName, userName, cookiesContent string) {
	GetUser(weiSiteName, userName).setCookies(cookiesContent)
}

func AddUserCookies(weiSiteName, userName, cookiesContent string, dbPrimaryKeyId int64) {
	user := GetUser(weiSiteName, userName)
	user.cookies = cookiesContent
	user.dbPrimaryKeyId = dbPrimaryKeyId
	user.saveCookies()
}
