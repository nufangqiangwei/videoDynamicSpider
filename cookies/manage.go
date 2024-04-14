package cookies

import (
	"strconv"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/log"
)

const (
	cookiesFileFolder = "baseCookies"
	blankUserName     = "空用户"
	Tourists          = "tourists"
)

var defaultUserCookies = map[string]string{
	"bilibili": "buvid3=6D9D1A89-B323-32AD-05D8-C0FA2F9E7E5709803infoc; b_nut=1712403109; i-wanna-go-back=-1; b_ut=7; b_lsid=7E44AA9E_18EB32DD7EF; _uuid=8818E256-AA95-9E7B-6173-B110E6315A77409879infoc; enable_web_push=DISABLE; FEED_LIVE_VERSION=V_HEADER_LIVE_NO_POP; header_theme_version=undefined; buvid4=2EC41739-8951-3D43-F7B7-D56CA6A0BE6110473-024040611-vuloyplW97cVGimTYVWpEw%3D%3D; buvid_fp=bef0fc320e7e23f4933b01d05e9e99f9; home_feed_column=5; browser_resolution=1920-959; CURRENT_FNVAL=4048; sid=mdpzjc3h; rpdid=|(uYmYY|||Jm0J'u~ukR|Jukl; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTI2NjIzMjksImlhdCI6MTcxMjQwMzA2OSwicGx0IjotMX0.kBJMrWxipe3kH6nm1uENI7mToN-NJOqCDX7yaiRtnx8; bili_ticket_expires=1712662269",
}

var DataSource baseStruct.CookiesFlush = privateReadLocalFile{}

type UserCookie struct {
	cookies              string
	lastFlushCookiesTime time.Time
	cookiesFail          bool
	fileName             string
	webSiteName          string
	dbPrimaryKeyId       int64 // 用户id
}

func (c *UserCookie) SetDBPrimaryKeyId(id int64) {
	c.dbPrimaryKeyId = id
}
func (c *UserCookie) GetDBPrimaryKeyId() int64 {
	if c.dbPrimaryKeyId == 0 {
		panic("未设置用户id")
	}
	return c.dbPrimaryKeyId
}

// GetStatus 获取cookies是否有效
func (c *UserCookie) GetStatus() bool {
	return !c.cookiesFail
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
	if c.cookiesFail || c.cookies == "" || c.lastFlushCookiesTime.Add(time.Hour*24).Before(time.Now()) {
		// 如果cookies失效并且上次刷新时间超过24小时，重新刷新cookies
		if c.cookiesFail {
			log.Info.Printf("%scookies标注为失效", c.fileName)
		} else if c.cookies == "" {
			log.Info.Printf("%scookies未加载", c.fileName)
		} else {
			log.Info.Printf("%scookies距离上次加载已过24小时。上次加载时间%s", c.fileName, c.lastFlushCookiesTime.Format("2006.01.02 15:04:05"))
		}
		c.lastFlushCookiesTime = time.Now()
		c.readFile()
		if c.cookiesFail {
			// cookies刷新失败
			if log.ErrorLog != nil {
				log.ErrorLog.Printf("cookies失效，请更新%scookies文件", c.fileName)
			} else {
				println("cookies失效，请更新", c.fileName, "cookies文件")
			}
			DataSource.UserCookiesInvalid(c.webSiteName, c.fileName, c.cookies, strconv.FormatInt(c.dbPrimaryKeyId, 10))
		}
	}
}

func (c *UserCookie) setCookies(cookiesContext string) {
	c.cookies = cookiesContext
}

func (c *UserCookie) saveCookies() {
	err := DataSource.UpdateUserCookies(c.webSiteName, c.fileName, c.cookies, strconv.FormatInt(c.dbPrimaryKeyId, 10))
	if err != nil {
		log.ErrorLog.Printf("持久化%scookies文件失败", c.fileName)
	}
}

func (c *UserCookie) readFile() {
	// 读取文件中的cookies
	if c.fileName == blankUserName {
		return
	}

	cookies := DataSource.GetUserCookies(c.webSiteName, c.fileName)
	if c.cookiesFail {
		c.cookiesFail = c.cookies == cookies
	}
	c.setCookies(cookies)
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

func NewDefaultUserCookie(webSiteName string) *UserCookie {
	cookie, ok := defaultUserCookies[webSiteName]
	if !ok {
		cookie = ""
	}
	return &UserCookie{
		fileName:    blankUserName,
		webSiteName: webSiteName,
		cookies:     cookie,
	}
}

func NewTemporaryUserCookie(webSiteName, cookiesText string) *UserCookie {
	return &UserCookie{
		cookies:     cookiesText,
		fileName:    blankUserName,
		webSiteName: webSiteName,
	}
}

type WebSiteCookies struct {
	webName string
	cookies []*UserCookie
	index   int
}

func (ws *WebSiteCookies) PickUser() *UserCookie {
	if ws.cookies == nil {
		panic("cookies尚未初始化")
	}
	if len(ws.cookies) == 0 {
		return NewDefaultUserCookie(ws.webName)
	}
	if ws.index == len(ws.cookies)-1 {
		ws.index = -1
	}
	ws.index++
	return ws.cookies[ws.index]
}

func (ws *WebSiteCookies) addUser(users []*UserCookie) {
	var have bool
	for _, u := range users {
		have = false
		for index, U := range ws.cookies {
			if u.fileName == U.fileName {
				ws.cookies[index] = U
				have = true
				break
			}
		}
		if !have {
			ws.cookies = append(ws.cookies, u)
		}
	}

}

type WebSiteCookiesManager struct {
	cookiesMap      map[string]*UserCookie
	touristsCookies []*UserCookie
	webSiteName     string
}

func (wcm *WebSiteCookiesManager) FlushCookies() {
	for _, userInfo := range DataSource.UserList(wcm.webSiteName) {
		c, ok := wcm.cookiesMap[userInfo.UserName]
		log.Info.Printf("加载%s网站%s用户cookies,cookies", wcm.webSiteName, userInfo.UserName)
		if !ok {
			c = &UserCookie{fileName: userInfo.UserName, webSiteName: wcm.webSiteName, cookies: userInfo.Content, lastFlushCookiesTime: time.Now()}
			wcm.cookiesMap[userInfo.UserName] = c
		}

	}
	wcm.GetTouristsCookies()
}

func (wcm *WebSiteCookiesManager) GetTouristsCookies() {
	cookiesList := DataSource.GetTouristsCookies(wcm.webSiteName)
	if cookiesList == nil || cap(cookiesList) == 0 {
		return
	}
	for _, cookie := range cookiesList {
		var isContinue bool
		isContinue = true
		for _, u := range wcm.touristsCookies {
			if u.cookies == cookie {
				isContinue = false
				break
			}
		}
		if isContinue {
			wcm.touristsCookies = append(wcm.touristsCookies, &UserCookie{
				fileName:             Tourists,
				webSiteName:          wcm.webSiteName,
				cookies:              cookie,
				lastFlushCookiesTime: time.Now(),
			})
		}
	}
}

var rootCookiesMap map[string]*WebSiteCookiesManager

func FlushAllCookies() {
	if rootCookiesMap == nil {
		rootCookiesMap = make(map[string]*WebSiteCookiesManager)
	}
	for _, webSiteName := range DataSource.WebSiteList() {
		w, ok := rootCookiesMap[webSiteName]
		if !ok {
			w = &WebSiteCookiesManager{
				cookiesMap:      make(map[string]*UserCookie),
				touristsCookies: make([]*UserCookie, 0),
				webSiteName:     webSiteName,
			}
			rootCookiesMap[webSiteName] = w
		}
		w.FlushCookies()
	}
}

func GetWeb(weiSiteName string) *WebSiteCookies {
	result := WebSiteCookies{cookies: make([]*UserCookie, 0), webName: weiSiteName}
	websiteData, ok := rootCookiesMap[weiSiteName]
	if !ok {
		return &result
	}
	var users []*UserCookie
	for _, u := range websiteData.cookiesMap {
		users = append(users, u)
	}
	result.addUser(users)
	result.index = -1
	return &result
}
func GetTouristsCookies(weiSiteName string) *WebSiteCookies {
	result := WebSiteCookies{cookies: make([]*UserCookie, 0), webName: weiSiteName}
	websiteData, ok := rootCookiesMap[weiSiteName]
	if !ok {
		return &result
	}
	result.addUser(websiteData.touristsCookies)
	result.index = -1
	return &result
}
func GetUser(weiSiteName, userName string) *UserCookie {
	w, ok := rootCookiesMap[weiSiteName]
	if !ok {
		w = &WebSiteCookiesManager{
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
	user.setCookies(cookiesContent)
	user.dbPrimaryKeyId = dbPrimaryKeyId
	user.saveCookies()
}
