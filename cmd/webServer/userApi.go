package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/mail"
	"strings"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

type registerUserRequestBody struct {
	Email    string `json:"email"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}

const emailRegexp = `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`

var (
	emailHostWhitelist = []string{"qq.com", "gmail.com", "163.com", "126.com", "sina.com.cn", "sohu.com"}
)

// 解析电子邮件地址
func extractHostFromEmail(email string) (string, error) {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}

	// 提取主机部分
	host := strings.Split(addr.Address, "@")[1]

	return host, nil
}
func registerUser(ctx *gin.Context) {
	body := registerUserRequestBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	if len(body.Password) > 18 {
		ctx.JSON(400, gin.H{"error": "密码长度不能超过18位"})
		return
	}
	// 获取email的host信息
	emailHost, err := extractHostFromEmail(body.Email)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid email"})
		return
	}
	println(emailHost)
	if emailHost == "" {
		ctx.JSON(400, gin.H{"error": "Invalid email"})
		return
	} else if !utils.InArray(emailHost, emailHostWhitelist) {
		ctx.JSON(400, gin.H{"error": "Email host not allowed"})
		return
	}
	user := models.User{
		UserName: body.UserName,
		Email:    body.Email,
		Password: body.Password,
	}
	transaction := models.GormDB.Begin()

	err = models.AddUser(user, transaction)
	if err != nil {
		transaction.Rollback()
		var userExist models.UserExist
		if errors.As(err, &userExist) {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		log.ErrorLog.Printf("注册用户失败:%s", err.Error())
		ctx.JSON(500, gin.H{"error": "数据库异常"})
		return
	}
	// 写入cookies
	cookies := generateUserCookies(user)
	if cookies == "" {
		transaction.Rollback()
		log.ErrorLog.Println("生成cookies失败:")
		ctx.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	transaction.Commit()
	ctx.SetCookie("info", cookies, ExpirationTime, "/", "", false, true)
	ctx.JSON(200, BaseResponse{
		Msg: "ok",
		Data: struct {
			UserName string `json:"userName"`
		}{UserName: user.UserName},
	})
}
func generateUserCookies(user models.User) string {
	// 生成cookie
	timeNow := utils.GetCurrentTime()
	token := Check{
		Token:          utils.GenerateRandomString(16),
		UserId:         user.ID,
		ExpirationTime: ExpirationTime,
		LoginTime:      timeNow,
	}
	err := models.UserLoginCache(user.ID, timeNow, ExpirationTime)
	if err != nil {
		log.ErrorLog.Println(err.Error())
		return ""
	}
	return utils.EncryptToken(token.String(), config.AesKey, config.AesIv)
}

type userLoginRequestBody struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

func userLogin(ctx *gin.Context) {
	body := userLoginRequestBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	user, err := models.CheckUser(body.UserName, body.Password)
	if err != nil {
		println(err.Error())
		ctx.JSON(400, gin.H{"error": "账号或密码错误"})
		return
	}
	// 写入cookies
	cookies := generateUserCookies(user)
	if cookies == "" {
		ctx.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	ctx.SetCookie("info", cookies, ExpirationTime, "/", "", false, true)
	ctx.JSON(200, BaseResponse{
		Msg: "ok",
		Data: struct {
			UserName string `json:"userName"`
		}{UserName: user.UserName},
	})
}

type resetUserPasswordRequestBody struct {
	UserName    string `json:"userName"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func resetUserPassword(ctx *gin.Context) {
	body := resetUserPasswordRequestBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	user := ctxGetUser(ctx)
	if user.UserName != body.UserName {
		ctx.JSON(403, gin.H{"error": "非法操作"})
		return
	}
	if user.CheckUserPassword(body.OldPassword) != nil {
		ctx.JSON(400, gin.H{"error": "原密码错误"})
		return
	}
	user.Password = body.NewPassword
	if user.SetPassword() != nil {
		ctx.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	ctx.JSON(200, BaseResponse{
		Msg: "ok",
		Data: struct {
			UserName string `json:"userName"`
		}{UserName: user.UserName},
	})
}

func ctxGetUser(ctx *gin.Context) models.User {
	var (
		u    any
		ok   bool
		user models.User
	)
	u, ok = ctx.Get("user")
	if !ok {
		ctx.JSON(411, gin.H{"error": "未登录"})
		ctx.Abort()
		return user
	}
	user, ok = u.(models.User)
	if !ok {
		ctx.JSON(411, gin.H{"error": "未登录"})
		ctx.Abort()
		return user
	}
	return user
}

type UserBindAuthor struct {
	WebName     string `json:"webName"`
	AuthorName  string `json:"authorName"`
	CookiesFail bool   `json:"cookiesFail"`
}

// 获取用户上传了cookies的网站信息
func getUserManageCookiesWebSite(ctx *gin.Context) {
	user := ctxGetUser(ctx)
	sql := `select w.web_name,a.author_name,ua.cookies_fail
from user_web_site_account ua 
    inner join author a on ua.author_id=a.id 
    inner join web_site w on w.id=a.web_site_id
         where ua.user_id=?`
	result := make([]UserBindAuthor, 0)
	err := models.GormDB.Raw(sql, user.ID).Scan(&result).Error
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	ctx.JSON(200, BaseResponse{
		Msg:  "ok",
		Data: result,
	})
}

type UserFollowAuthor struct {
	WebName        string `json:"webName"`
	UserAuthorId   int64  `json:"userAuthorId"`
	UserAuthorName string `json:"userAuthorName"`
	UserAuthorUid  string `json:"userAuthorUid"`
	AuthorName     string `json:"authorName"`
	AuthorAvatar   string `json:"authorAvatar"`
	AuthorDesc     string `json:"authorDesc"`
	AuthorUid      string `json:"authorUid"`
	AuthorId       int64  `json:"authorId"`
}

//获取平台用户关联的账号，所有关注的视频主
func getUserFollowAuthor(ctx *gin.Context) {
	user := ctxGetUser(ctx)
	sql := `select w.web_name,
       a.id,
       a.author_web_uid,
       a.author_name,
       ab.author_name,
       ab.avatar,
       ab.author_desc,
       ab.author_web_uid,
       ab.id
from user_web_site_account ua
         inner join author a on ua.author_id = a.id
         inner join web_site w on w.id = a.web_site_id
         inner join follow f on f.user_id = ua.author_id
         inner join author ab on ab.id = f.author_id
where ua.user_id = ? limit 10`
	result := make([]UserFollowAuthor, 0)
	err := models.GormDB.Raw(sql, user.ID).Scan(&result).Error
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	ctx.JSON(200, BaseResponse{
		Msg:  "ok",
		Data: result,
	})
}

type uploadWebCookiesRequestBody struct {
	WebName string `json:"webName"`
	Cookies string `json:"cookies"`
}

// 用户上传自己的cookies
func uploadWebCookies(ctx *gin.Context) {
	body := uploadWebCookiesRequestBody{}
	err := ctx.BindJSON(&body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	var webSpider models.VideoCollection
	for _, spider := range spiderManager.collection {
		if spider.GetWebSiteName().WebName == body.WebName {
			webSpider = spider
			break
		}
	}
	if webSpider == nil {
		ctx.JSON(503, gin.H{"error": "尚未支持的网站"})
		return
	}
	accountInfo := webSpider.GetSelfInfo(body.Cookies)
	if accountInfo == nil {
		ctx.JSON(400, gin.H{"error": "cookies无效"})
		return
	}
	accountName := accountInfo.AccountName()
	a := models.Author{}
	err = models.GormDB.Model(&models.Author{}).Where("author_name=? and web_site.web_name=?", accountName, body.WebName).Joins(
		"web_site on web_site.id=author.web_site_id",
	).Scan(&a).Error
	if err != nil {
		ctx.JSON(500, gin.H{"error": "数据库查询失败"})
		return
	}
	if a.Id == 0 {
		a = accountInfo.GetAuthorModel()
		err = models.GormDB.Create(&a).Error
		if err != nil {
			ctx.JSON(500, gin.H{"error": "数据库查询失败"})
			return
		}
	}
	user := ctxGetUser(ctx)
	userBindAccount := models.UserWebSiteAccount{}
	models.GormDB.Model(&models.UserWebSiteAccount{}).Where("user_id=? and author_id=?", user.ID, a.Id).Scan(&userBindAccount)
	if userBindAccount.ID == 0 {
		userBindAccount = models.UserWebSiteAccount{
			UserId:   user.ID,
			AuthorId: a.Id,
		}
		models.GormDB.Create(&userBindAccount)
	}
	cookies.AddUserCookies(body.WebName, a.AuthorName, body.Cookies, a.Id)
	ctx.JSON(200, successResponse)
}
