package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/mail"
	"strings"
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
		utils.ErrorLog.Printf("注册用户失败:%s", err.Error())
		ctx.JSON(500, gin.H{"error": "数据库异常"})
		return
	}
	// 写入cookies
	cookies := generateUserCookies(user)
	if cookies == "" {
		transaction.Rollback()
		utils.ErrorLog.Println("生成cookies失败:")
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
		utils.ErrorLog.Println(err.Error())
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
	var (
		u    any
		ok   bool
		user models.User
	)
	u, ok = ctx.Get("user")
	if !ok {
		ctx.JSON(403, gin.H{"error": "未登录"})
		return
	}
	user, ok = u.(models.User)
	if !ok {
		ctx.JSON(403, gin.H{"error": "未登录"})
		return
	}
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
