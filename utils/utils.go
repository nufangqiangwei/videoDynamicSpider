package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
	"io"
	"os"
	"path"
	"time"
)

const maxFileSize = 100 * 1024 * 1024

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

type WriteFile struct {
	FolderPrefix     []string
	FileNamePrefix   string
	file             *os.File
	writeNumber      int
	lastOpenFileName string
}

func (wf *WriteFile) getFileName() string {
	if wf.lastOpenFileName == "" {
		return fmt.Sprintf("%s-%s.json", wf.FileNamePrefix, time.Now().Format("2006-01-02-15-04-05"))
	}
	return wf.lastOpenFileName

}
func (wf *WriteFile) checkFileSize() {
	if wf.file == nil {
		filePath := append(wf.FolderPrefix, wf.getFileName())
		f, err := os.OpenFile(path.Join(filePath...), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			ErrorLog.Printf("打开新文件失败%s", err.Error())
			panic(err)
		}
		wf.file = f
		wf.lastOpenFileName = ""
		return
	}
	for {
		fi, err := wf.file.Stat()
		if err != nil {
			ErrorLog.Printf("获取文件信息失败%s", err.Error())
			panic(err)
		}
		if fi.Size() >= maxFileSize {
			wf.file.Close()
			filePath := append(wf.FolderPrefix, wf.getFileName())
			f, err := os.OpenFile(path.Join(filePath...), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				ErrorLog.Printf("打开新文件失败%s", err.Error())
				panic(err)
			}
			wf.file = f
			wf.writeNumber = 0
		} else {
			break
		}
	}
	wf.lastOpenFileName = ""
}
func (wf *WriteFile) Write(data []byte) (int, error) {
	if wf.file == nil {
		wf.checkFileSize()
	}
	// 每写入两千行就检查下文件大小
	if wf.writeNumber%2000 == 0 {
		wf.checkFileSize()
	}
	wf.writeNumber++
	return wf.file.Write(data)
}
func (wf *WriteFile) WriteLine(data []byte) (int, error) {
	a, b := wf.Write(data)
	if b != nil {
		return a, b
	}
	return wf.Write([]byte{10})
}
func (wf *WriteFile) Close() {
	wf.lastOpenFileName = wf.file.Name()
	wf.file.Close()
}
