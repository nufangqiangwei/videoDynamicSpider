package models

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID       int64  `json:"id" gorm:"primary_key"`
	UserName string `json:"userName" gorm:"size:255;uniqueIndex:username"`
	Password string `json:"password"`
	Email    string `json:"email" gorm:"size:255;uniqueIndex:email"`
	Prohibit bool   `json:"prohibit" gorm:"default:false"`
}

// 生成密码的哈希值
func hashPassword(password string) (string, error) {
	print(len([]byte(password)))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// 比较哈希值和密码是否匹配
func comparePasswords(hashedPassword, password string) error {
	print(len([]byte(password)))
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

type UserExist struct {
	userName string
}

func (u UserExist) Error() string {
	return fmt.Sprintf("用户 %s 已存在", u.userName)
}

type UserNotExist struct {
	userName string
}

func (u UserNotExist) Error() string {
	return fmt.Sprintf("用户 %s 不存在", u.userName)
}

func AddUser(user User, transaction *gorm.DB) error {
	var (
		err error
		u   User
	)
	u = User{}

	tx := transaction.Model(&User{}).Where("user_name = ?", user.UserName).First(&u)
	if tx.Error != nil && errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		println(tx.Error.Error())
		return UserExist{userName: user.UserName}
	}

	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return err
	}
	err = transaction.Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func CheckUser(username, password string) (User, error) {
	var (
		err error
		u   User
	)
	u = User{}
	tx := GormDB.Model(&User{}).Where("user_name = ?", username).First(&u)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return u, UserNotExist{username}
		}
		return u, tx.Error
	}
	err = comparePasswords(u.Password, password)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (user User) CheckUserPassword(password string) error {
	return comparePasswords(user.Password, password)
}
func (user User) SetPassword() error {
	var err error
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return err
	}
	return GormDB.Model(&User{}).Where("id = ?", user.ID).Update("password", user.Password).Error
}
