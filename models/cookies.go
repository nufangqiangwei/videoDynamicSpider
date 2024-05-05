package models

import (
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/baseStruct"
)

type UserCookies struct {
	ID         int64     `json:"id" gorm:"primary_key"`
	WebSiteId  int64     `json:"webSiteId" gorm:"index:web_site_id"`
	UserId     int64     `json:"userId" gorm:"index:user_id"` // UserId和AuthorId值为0的时候，代表这个是游客cookies
	AuthorId   int64     `json:"authorId" gorm:"index:author_id"`
	Content    string    `json:"-" gorm:"text"`
	UpdateTime time.Time `json:"updateTime" gorm:"default:CURRENT_TIMESTAMP"`
	Spider     int       `json:"spider" gorm:"default:0"` // 指定哪个爬虫可以读取，0代表所有的都可以读取
	Valid      bool      `json:"valid"`                   // 1有效 0无效
}

// buvid3=7F8A14F9-F832-F87C-EE15-1A58BD40FE4E21983infoc; b_nut=1713160521; b_lsid=68ADECEE_18EE0530A4E; _uuid=10451C79A-1FC2-12D9-10691-2C1B1091106621022338infoc; enable_web_push=DISABLE; FEED_LIVE_VERSION=V_WATCHLATER_PIP_WINDOW3; header_theme_version=undefined; buvid_fp=2e9493d48316cab2dcc8f0eda65313b6; buvid4=D8D53FA4-4E32-B2F5-3939-356171CC2C8922836-024041505-03dNWZWH1CFEqbJsdYgv7w%3D%3D; home_feed_column=5; browser_resolution=1920-957; SESSDATA=313e061b%2C1728712540%2Ced87b%2A42CjAC6frfS-MRS8Cc_vmkA8zHVFeHBPVbvFSUfcBAalwTeCoz1iwyKJ4jRgr_cdWlB1oSVjBISXZ2SXFiRHdNQW40R2xYRUtHWFZWR1FpUEc5YllRUUVrTkFZWTZGREMzcTYzbTd6RlJNaXc4d2dJZS15MWdYbGlBVjgtMnlsT29BSUctMk9xQUJBIIEC; bili_jct=ede61e080c83b834db973b14ffb9fd61; DedeUserID=10932398; DedeUserID__ckMd5=d5bdc808e48784b2; sid=7z9u6o8u; CURRENT_FNVAL=4048; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTM0MTk3NDYsImlhdCI6MTcxMzE2MDQ4NiwicGx0IjotMX0.ZOcqw1YC3CneW2bqI-u4fYbIe7wk23sVHF0sq97ctdo; bili_ticket_expires=1713419686; rpdid=|(J|~)~u)kl)0J'u~uJuY|lR)
// LIVE_BUVID=AUTO6015975742804240; i-wanna-go-back=-1; CURRENT_BLACKGAP=0; buvid_fp_plain=undefined; b_timer=%7B%22ffp%22%3A%7B%22333.1007.fp.risk_4609B7F9%22%3A%22182B928B2A2%22%2C%22333.788.fp.risk_4609B7F9%22%3A%22182B9289B66%22%2C%22333.880.fp.risk_4609B7F9%22%3A%221811447FB84%22%2C%22333.967.fp.risk_4609B7F9%22%3A%221816C2A2C0C%22%2C%22444.8.fp.risk_4609B7F9%22%3A%221825ED57D31%22%2C%22333.976.fp.risk_4609B7F9%22%3A%22182785823B2%22%2C%22888.2421.fp.risk_4609B7F9%22%3A%22180F4EBB0E9%22%2C%22444.7.fp.risk_4609B7F9%22%3A%221825F17A69A%22%2C%22444.13.fp.risk_4609B7F9%22%3A%22181B36A01FF%22%2C%22777.5.0.0.fp.risk_4609B7F9%22%3A%22180F4EBAACE%22%2C%22333.337.fp.risk_4609B7F9%22%3A%221829AE2AF70%22%2C%22333.999.fp.risk_4609B7F9%22%3A%221829AE2D3DC%22%2C%22444.55.fp.risk_4609B7F9%22%3A%221816C4F28E3%22%2C%22333.905.fp.risk_4609B7F9%22%3A%22182C61C93EE%22%2C%22333.1073.fp.risk_4609B7F9%22%3A%2218251E7956A%22%2C%22333.937.fp.risk_4609B7F9%22%3A%221825F14038A%22%2C%22888.70696.fp.risk_4609B7F9%22%3A%221825F16DC22%22%2C%22333.1193.fp.risk_4609B7F9%22%3A%22182B66389F4%22%7D%7D; header_theme_version=CLOSE; theme_style=light; FEED_LIVE_VERSION=V8; hit-new-style-dyn=1; hit-dyn-v2=1; DedeUserID=10932398; DedeUserID__ckMd5=d5bdc808e48784b2; b_ut=5; go-back-dyn=0; buvid3=FE124433-BDEA-4CC0-B3F1-E5A3DD37F23292667infoc; b_nut=1692186191; _uuid=76D551087-3E32-4E12-46FE-E6B410E1CC62733545infoc; buvid4=1EC43764-D911-026E-04BC-A666638B1AE237315-022012118-dJ30JF9%2BylKXU%2FQmcpKHPA%3D%3D; i-wanna-go-feeds=2; kfcSource=cps_comments_4624257; msource=cps_comments_4624257; deviceFingerprint=0981a6f5338d4165063a8c2ad4f98796; enable_web_push=DISABLE; CURRENT_QUALITY=80; rpdid=0zbfVGePcl|1a5eQ1fzL|15|3w1RgDx9; share_source_origin=COPY; fingerprint=bef0fc320e7e23f4933b01d05e9e99f9; home_feed_column=5; CURRENT_FNVAL=4048; bsource=search_google; buvid_fp=bef0fc320e7e23f4933b01d05e9e99f9; SESSDATA=cd9ef0be%2C1728736737%2Cb28c4%2A42CjAzGpj78Yc57IDVT7bipYArB2j1_WIDRxQZHIPgtlpYsbeUCx_VkioLpJu1sTv5XKkSVnAzRW9nWm9BUEc1dldZS3J1eVBkcGRVY3RXUEQzM0c5UWNGa1hrcDBpVk5MZXBiLVhNeE5lbkJ4VmFYMzAzUUJLa0tSemJMMF9fckh3M1ljY3lzdGl3IIEC; bili_jct=b206a5e2d3e93df43ddd9bb1a8155d6d; sid=6hqli35c; bp_video_offset_10932398=920611679992021011; PVID=1; b_lsid=95163EEC_18EEC8AA46B; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTM2MjQ3MDQsImlhdCI6MTcxMzM2NTQ0NCwicGx0IjotMX0.1vYABW-mbTtSUC27O-MXi0zO-pNG0rH4itBeDaH25QI; bili_ticket_expires=1713624644; browser_resolution=1920-959
// WebSiteCookies 通过数据库实现 baseStruct.CookiesFlush 接口
type WebSiteCookies struct {
	Spider int
}

func (wsc WebSiteCookies) WebSiteList() []string {
	result := make([]string, 0)
	GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id").Where("spider=?", wsc.Spider).Group("web_site.web_name").Pluck("web_site.web_name", &result)
	return result
}
func (wsc WebSiteCookies) UserList(webName string) []baseStruct.CacheUserCookies {
	result := make([]baseStruct.CacheUserCookies, 0)
	GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webName).Joins("inner join author on author.id=author_id").Where("spider=?", wsc.Spider).Select("author.author_name as user_name,content").Scan(&result)
	return result
}
func (wsc WebSiteCookies) GetUserCookies(webSiteName, userName string) string {
	var result UserCookies
	GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webSiteName).Joins("inner join author on author.id=author_id and author.author_name=?", userName).Where("spider=?", wsc.Spider).First(&result)
	return ""
}
func (wsc WebSiteCookies) UpdateUserCookies(webSiteName, authorName, cookiesContent, userId string) error {
	// 判断这条数据是否存在，不存在插入数据，存在才更新
	var (
		result UserCookies
		tx     *gorm.DB
	)
	tx = GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webSiteName).Joins("inner join author on author.id=author_id and author.author_name=?", authorName).First(&result)
	if tx.Error != nil {
		return tx.Error
	}
	if result.ID == 0 {
		// 插入数据，在插入数据前需要判断web_site_id、author_id、userId是否存在
		var webSite WebSite
		tx = GormDB.Model(&WebSite{}).Where("web_name=?", webSiteName).First(&webSite)
		if tx.Error != nil {
			return tx.Error
		}
		if webSite.Id == 0 {
			return NewWebSiteNotExist(webSiteName)
		}
		var author Author
		tx = GormDB.Model(&Author{}).Where("author_name=?", authorName).First(&author)
		if tx.Error != nil {
			return tx.Error
		}
		if author.Id == 0 {
			return NewAuthorNotExist(authorName)
		}
		var user User
		tx = GormDB.Model(&User{}).Where("id=?", userId).First(&user)
		if tx.Error != nil {
			return tx.Error
		}
		if user.ID == 0 {
			return NewUserNotExist(userId)
		}
		tx = GormDB.Create(&UserCookies{WebSiteId: webSite.Id, UserId: user.ID, AuthorId: author.Id, Content: cookiesContent, Spider: wsc.Spider})
		return tx.Error
	}
	tx = GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webSiteName).Joins("inner join author on author.id=author_id and author.author_name=?", authorName).Where("spider=?", wsc.Spider).Update("content", cookiesContent)
	return tx.Error
}
func (wsc WebSiteCookies) UserCookiesInvalid(webSiteName, authorName, cookiesContent, userId string) error {
	return nil
}
func (wsc WebSiteCookies) GetTouristsCookies(webSiteName string) []string {
	touristsCookiesList := []string{}
	GormDB.Model(&UserCookies{}).Joins("inner join web_site on web_site.id=web_site_id and web_site.web_name=?", webSiteName).Where("spider=? and user_id=0 and author_id=0 and valid=1", wsc.Spider).Pluck("content", &touristsCookiesList)
	return touristsCookiesList
}
