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
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
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

	filePath := "E:\\GoCode\\videoDynamicAcquisition\\allVideo"
	fileNameList := []string{}
	filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "json") {
			fileNameList = append(fileNameList, info.Name())
		}
		return nil
	})
	responseStruct := new(bilibili.VideoListPageResponse)
	utils.InitLog(baseStruct.RootPath)
	models.InitDB("spider:spider@tcp(100.124.177.135:3306)/videoSpider?charset=utf8mb4&parseTime=True&loc=Local")
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
		authorVideoUUIDList := []string{}
		lastAuthorMid := 0
		var authorId int64 = 0
		for scanner.Scan() {
			line := scanner.Bytes()
			err = json.Unmarshal(line, responseStruct)
			if err != nil {
				continue
			}
			authorMid := responseStruct.Data.List.Vlist[0].Mid
			if lastAuthorMid != authorMid {
				// 查询这个作者本地保存的视频信息
				// select v.uuid from video v inner join author a on v.author_id = a.id where a.author_web_uid= 1635;
				models.GormDB.Table("video v").
					Select("v.uuid").
					Joins("inner join author a on v.author_id = a.id").
					Where("a.author_web_uid = ?", authorMid).
					Find(&authorVideoUUIDList)
				// 查询这个作者的id
				// select id from author where author_web_uid= 1635;
				models.GormDB.Table("author").
					Select("id").
					Where("author_web_uid = ?", authorMid).
					Find(&authorId)
				lastAuthorMid = authorMid
			}
			for _, videoInfo := range responseStruct.Data.List.Vlist {
				if !utils.InArray(videoInfo.Bvid, authorVideoUUIDList) {
					createdTime := time.Unix(videoInfo.Created, 0)
					// 保存视频信息
					vv := models.Video{
						WebSiteId:  1,
						Title:      videoInfo.Title,
						VideoDesc:  videoInfo.Description,
						Duration:   bilibili.HourAndMinutesAndSecondsToSeconds(videoInfo.Length),
						Uuid:       videoInfo.Bvid,
						Url:        "",
						CoverUrl:   videoInfo.Pic,
						UploadTime: &createdTime,
						CreateTime: time.Now(),
					}
					vv.Save()
					authorVideoUUIDList = append(authorVideoUUIDList, videoInfo.Bvid)
				}
			}

		}
		err = scanner.Err()
		if err != nil {
			println(err.Error())
		}
		testIndex++
	}
}

func TestPrefixByte(t *testing.T) {
	// {"url":"","response":}
	//prefix := "{\"url\":\""
	//brackets := "}"
	//suffix := "\",\"response\":"
	//fmt.Printf("%v\n", []byte(prefix))
	//fmt.Printf("%v\n", []byte(brackets))
	//fmt.Printf("%v\n", []byte(suffix))
	prefixByte := []byte{123, 34, 117, 114, 108, 34, 58, 34}
	bracketsByte := []byte{125}
	suffixByte := []byte{34, 44, 34, 114, 101, 115, 112, 111, 110, 115, 101, 34, 58}
	responseByte := []byte(`{"code":0,"msg":"ok"}`)
	urlByte := []byte(`https://api.bilibili.com/x/space/arc/search?mid=1635&ps=30&tid=0&pn=1&keyword=&order=pubdate&jsonp=jsonp`)
	print(string(prefixByte))
	print(string(urlByte))
	print(string(suffixByte))
	print(string(responseByte))
	println(string(bracketsByte))
}
