package main

import (
	"github.com/google/uuid"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
	"videoDynamicAcquisition/baseStruct"
)

func TestUUID(t *testing.T) {
	println(uuid.NewString())
	println(uuid.NewString())
	println(uuid.NewString())
	println(uuid.NewString())
}

func TestCreateFolder(t *testing.T) {
	createFolder(false, "allVideo")
	taskId := "f4c37262-c9e3-4e38-8717-7962ca7dfc79"
	err := createFolder(true, "allVideo", taskId)
	for err != nil {
		if err.Error() == "文件夹已存在" {
			println("重新创建文件夹")
			taskId = uuid.NewString()
			err = createFolder(true, "allVideo", taskId)
		} else {
			println("创建文件夹失败")
			return
		}
	}
}
func TestWriteFile(t *testing.T) {
	taskId := "f4c37262-c9e3-4e38-8717-7962ca7dfc79"
	file := WriteFile{
		folderPrefix:   []string{baseStruct.RootPath, "allVideo", taskId},
		fileNamePrefix: "allVideo",
	}
	file.checkFileSize()
	file.file.Write([]byte{1, 3, 4, 67})
	file.checkFileSize()
	file.file.Write([]byte{1, 3, 4, 67})
	file.checkFileSize()
}
func TestWriteRequestParams(t *testing.T) {
	IdList := []string{
		"31728620",
		"31791052",
		"31814567",
		"32078085",
		"32123905",
	}
	taskId := "f4c37262-c9e3-4e38-8717-7962ca7dfc79"
	fileName := path.Join(baseStruct.RootPath, "allVideo", taskId, "requestParams")
	slice2 := make([]string, len(IdList))
	copy(slice2, IdList)
	os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
	slice2 = slice2[1:]
	os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
	slice2 = slice2[1:]
	os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
	slice2 = slice2[1:]
	os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)

}
