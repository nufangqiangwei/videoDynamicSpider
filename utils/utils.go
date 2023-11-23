package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
	"io"
	"os"
)

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

func IsUniqueErr(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			return true
		}
	}
	return false
}

func IsMysqlUniqueErr(err error) bool {
	var mysqlErr *mysql.MySQLError
	mysqlErr = new(mysql.MySQLError)
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

type DBFileLock struct {
	S string
}

// 计算文件的md5值
func GetFileMd5(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		ErrorLog.Printf("无法打开文件:%s, 错误信息：%s", filePath, err.Error())
		return "", err
	}
	defer file.Close()
	// 创建一个MD5哈希对象
	hash := md5.New()
	// 将文件内容拷贝到哈希对象中
	if _, err := io.Copy(hash, file); err != nil {
		ErrorLog.Printf("无法拷贝文件内容:%s, 错误信息：%s", filePath, err.Error())
		return "", err
	}
	// 计算MD5值
	md5Hash := hash.Sum(nil)
	// 将MD5值转换为字符串
	md5Str := hex.EncodeToString(md5Hash)
	return md5Str, nil
}
