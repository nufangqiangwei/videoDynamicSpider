package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
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
func TestIntoFileData(t *testing.T) {
	//
	filePath := "C:\\Code\\GO\\videoDynamicSpider\\cmd\\spiderProxy\\allVideo"
	fileNameList := []string{}
	filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileNameList = append(fileNameList, info.Name())
		}
		return nil
	})

	responseStruct := new(bilibili.VideoListPageResponse)
	db := baseStruct.CanUserDb()
	lineValues := []string{}
	sqlFile, _ := os.OpenFile(path.Join(filePath, "videoInertInto.sql"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	testIndex := 0
	for _, fileName := range fileNameList {
		fmt.Printf("%s\n", fileName)
		if testIndex >= 10 {
			return
		}
		file, err := os.Open(path.Join(filePath, fileName))

		if err != nil {
			println(err.Error())
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Bytes()
			err = json.Unmarshal(line, responseStruct)
			if err != nil {
				continue
			}

			x := saveAuthorAllVideo(db, *responseStruct, 1)
			if x != nil {
				lineValues = append(lineValues, x...)
			}
			if len(lineValues) > 100 {
				sqlFile.WriteString(fmt.Sprintf("insert into video (web_site_id, author_id, title, video_desc, duration, uuid, cover_url, upload_time)  values %s;\n", strings.Join(lineValues, ",")))
				lineValues = []string{}
				testIndex++
			}
		}
		err = scanner.Err()
		if err != nil {
			println(err.Error())
		}

	}
}
