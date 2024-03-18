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
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
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

	filePath := "C:\\Code\\GO\\videoDynamicSpider\\cmd\\webServer\\allVideo"
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
	log.InitLog(baseStruct.RootPath)
	models.InitDB("spider:spider@tcp(192.168.1.25:3306)/videoSpider?charset=utf8mb4&parseTime=True&loc=Local", false, nil)
	testIndex := 0
	nowTime := time.Now()
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
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		authorVideoUUIDList := []models.Video{}
		lastAuthorMid := 0
		var authorId int64 = 0
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			if line[len(line)-1] == 44 {
				line = line[:len(line)-1]
			}
			err = json.Unmarshal(line, responseStruct)
			if err != nil {
				println(err.Error())
				continue
			}
			if len(responseStruct.Data.List.Vlist) == 0 {
				continue
			}
			authorMid := responseStruct.Data.List.Vlist[0].Mid
			if lastAuthorMid != authorMid {
				// 查询这个作者本地保存的视频信息
				// select v.uuid from video v inner join author a on v.author_id = a.id where a.author_web_uid= 1635;
				models.GormDB.Table("video v").
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
				println(responseStruct.Data.List.Vlist[0].Author)
			}
			for _, videoInfo := range responseStruct.Data.List.Vlist {
				have := false
				for _, mysqlVideo := range authorVideoUUIDList {
					if mysqlVideo.Uuid == videoInfo.Bvid {
						createdTime := time.Unix(videoInfo.Created, 0)
						if mysqlVideo.CreateTime.IsZero() {
							mysqlVideo.CreateTime = nowTime
						}
						models.GormDB.Model(&mysqlVideo).Updates(map[string]interface{}{
							"upload_time": createdTime,
							"video_desc":  videoInfo.Description,
							"duration":    bilibili.HourAndMinutesAndSecondsToSeconds(videoInfo.Length),
							"cover_url":   videoInfo.Pic,
							"create_time": mysqlVideo.CreateTime,
						})
						have = true
						break
					}
				}
				if !have {
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
					vv.UpdateVideo()
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
	requestUrl := `https://api.bilibili.com/x/space/arc/search?mid=1635&ps=30&tid=0&pn=1&keyword=&order=pubdate&jsonp=jsonp`
	prefixByte := []byte{123, 34, 117, 114, 108, 34, 58, 34}
	bracketsByte := []byte{125}
	suffixByte := []byte{34, 44, 34, 114, 101, 115, 112, 111, 110, 115, 101, 34, 58}
	responseByte := []byte(`{"code":0,"msg":"ok"}`)
	urlByte := []byte(requestUrl)
	print(string(prefixByte))
	print(string(urlByte))
	print(string(suffixByte))
	print(string(responseByte))
	println(string(bracketsByte))
	println()
	println(string(writeRequestUrl(requestUrl, responseByte)))
}

func TestStr(t *testing.T) {
	xxx := "Go语言的字符有以下两种："
	yyy := []byte(xxx)
	fmt.Printf("%v\n", yyy)
	yyy = []byte("{}\n")
	fmt.Printf("%v\n", yyy)
}

type QueryData struct {
	WebName     string `json:"webName"`
	AuthorName  string `json:"authorName"`
	CookiesFail bool   `json:"cookiesFail"`
}

func TestModelQueryBindStruct(t *testing.T) {
	err := readConfig()
	if err != nil {
		println(err.Error())
		os.Exit(4)
	}
	logBlockList := log.InitLog(baseStruct.RootPath, "database")
	var databaseLog log.LogInputFile
	for _, logBlock := range logBlockList {
		if logBlock.FileName == "database" {
			databaseLog = logBlock
			break
		}
	}
	models.InitDB("spider:p0o9i8u7@tcp(database:3306)/video?charset=utf8mb4&parseTime=True&loc=Local", false, databaseLog.WriterObject)
	sql := `select w.web_name,a.author_name,ua.cookies_fail
from user_web_site_account ua 
    inner join author a on ua.author_id=a.id 
    inner join web_site w on w.id=a.web_site_id
         where ua.user_id=?`
	result := []QueryData{}
	err = models.GormDB.Raw(sql, 11).Scan(&result).Error
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("%v\n", result)
}
