package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"time"
	"videoDynamicAcquisition/log"
)

const maxFileSize = 100 * 1024 * 1024 // 100M

func InArray[T string | int64](val T, array []T) bool {
	for _, v := range array {
		if v == val {
			return true
		}
	}
	return false
}

// ArrayDifference 在slice1但是不在slice2的值
func ArrayDifference[T string | int64](slice1, slice2 []T) []T {
	m := make(map[T]T)
	for _, v := range slice1 {
		m[v] = v
	}
	for _, v := range slice2 {
		if _, ok := m[v]; ok {
			delete(m, v)
		}
	}
	var str []T

	for _, s2 := range m {
		str = append(str, s2)
	}
	return str
}

// ArrayDifferenceByStruct 在slice1但是不在slice2的值
//func ArrayDifferenceByStruct[T models.VideoAuthor | models.VideoTag](slice1, slice2 []T, keyFunc func(T) string) []T {
//	m := make(map[string]T)
//	for _, v := range slice1 {
//		m[keyFunc(v)] = v
//	}
//	for _, v := range slice2 {
//		if _, ok := m[keyFunc(v)]; ok {
//			delete(m, keyFunc(v))
//		}
//	}
//	var str []T
//
//	for _, s2 := range m {
//		str = append(str, s2)
//	}
//	return str
//
//}

//func IsUniqueErr(err error) bool {
//	var sqliteErr sqlite3.Error
//	if errors.As(err, &sqliteErr) {
//		if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique ||
//			sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
//			return true
//		}
//	}
//	return false
//}

type DBFileLock struct {
	S string
}

// 计算文件的md5值
func GetFileMd5(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		log.ErrorLog.Printf("无法打开文件:%s, 错误信息：%s", filePath, err.Error())
		return "", err
	}
	defer file.Close()
	// 创建一个MD5哈希对象
	hash := md5.New()
	// 将文件内容拷贝到哈希对象中
	if _, err := io.Copy(hash, file); err != nil {
		log.ErrorLog.Printf("无法拷贝文件内容:%s, 错误信息：%s", filePath, err.Error())
		return "", err
	}
	// 计算MD5值
	md5Hash := hash.Sum(nil)
	// 将MD5值转换为字符串
	md5Str := hex.EncodeToString(md5Hash)
	return md5Str, nil
}

func CheckFileWriteStatus(filePath string) bool {
	// 检查文件是否正在写入
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return false
	}
	defer file.Close()
	_, err = file.Seek(0, 2)
	if err != nil {
		return false
	}
	return true
}

func GetCurrentTime() int64 {
	return time.Now().Unix()
}

func GenerateRandomString(length int) string {
	// 定义包含所有可能字符的字符集
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 创建一个字节数组，用于存储生成的随机字符
	randomString := make([]byte, length)

	// 从字符集中随机选择字符，并将其添加到随机字符串中
	for i := 0; i < length; i++ {
		randomString[i] = charSet[rand.Intn(len(charSet))]
	}

	// 将字节数组转换为字符串并返回
	return string(randomString)
}
