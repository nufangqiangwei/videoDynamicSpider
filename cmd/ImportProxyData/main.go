package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

type void struct{}

var (
	member           void
	config           *utils.Config
	waitImportPath   string
	importingPath    string
	finishImportPath string
)

const (
	sleepTime = time.Minute
)

func readConfig() error {
	fileData, err := os.ReadFile("C:\\Code\\GO\\videoDynamicSpider\\cmd\\ImportProxyData\\config.json")
	if err != nil {
		println(err.Error())
		return err
	}
	config = &utils.Config{}
	err = json.Unmarshal(fileData, config)
	if err != nil {
		println(err.Error())
		return err
	}
	fmt.Printf("%v\n", *config)
	return nil
}
func init() {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		println("时区设置错误")
		os.Exit(2)
		return
	}
	time.Local = location
	err = readConfig()
	if err != nil {
		os.Exit(2)
		return
	}
	baseStruct.RootPath = config.ProxyDataRootPath
	utils.InitLog(baseStruct.RootPath)

	models.InitDB(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DB.User, config.DB.Password, config.DB.HOST, config.DB.Port, config.DB.DatabaseName))

}

func main() {
	// 死循环读取 waitImportFile这个目录下的文件，有新的文件出现就将这个文件移动到importingFile目录下，然后开始导入数据
	// 导入完成后，将这个文件移动到finishImportFile目录下
	waitImportPath = path.Join(baseStruct.RootPath, utils.WaitImportPrefix)
	importingPath = path.Join(baseStruct.RootPath, utils.ImportingPrefix)
	finishImportPath = path.Join(baseStruct.RootPath, utils.FinishImportPrefix)
	//for {
	// 检查importingPath目录下的文件是否有上次异常退出残留下来的文件
	importingFileList, err := os.ReadDir(importingPath)
	if err != nil {
		utils.ErrorLog.Printf("读取目录失败：%s\n", err.Error())
	} else {
		for _, importingFile := range importingFileList {
			if strings.HasSuffix(importingFile.Name(), "tar.gz") {
				importFileData(importingFile.Name())
			}
		}
	}
	// 读取waitImportPath目录下的文件
	waitImportFileList, err := os.ReadDir(waitImportPath)
	if err != nil {
		utils.ErrorLog.Printf("读取目录失败：%s\n", err.Error())
		//time.Sleep(sleepTime)
		//continue
		return
	}
	for _, waitImportFile := range waitImportFileList {
		if strings.HasSuffix(waitImportFile.Name(), "tar.gz") {
			// 检查文件是否正在写入
			if !utils.CheckFileWriteStatus(path.Join(waitImportPath, waitImportFile.Name())) {
				continue
			}
			// 将这个文件移动到importingPath目录下
			err = os.Rename(path.Join(waitImportPath, waitImportFile.Name()), path.Join(importingPath, waitImportFile.Name()))
			if err != nil {
				utils.ErrorLog.Printf("移动文件失败：%s\n", err.Error())
				time.Sleep(sleepTime)
				continue
			}
			// 开始导入数据
			importFileData(waitImportFile.Name())
		}
	}
	//time.Sleep(sleepTime)
	//}
}

func updateBilibiliVideoDetailInfo(response bilibili.VideoDetailResponse, WebSiteId int64) {
	video := models.Video{}
	var tx *gorm.DB
	tx = models.GormDB.Where("uuid = ?", response.Data.View.Bvid).Preload("Authors").Preload("Tag").
		Limit(1).Find(&video)
	if tx.Error != nil {
		utils.ErrorLog.Printf("获取视频信息失败：%s\n", tx.Error.Error())
		return
	}
	if video.Id == 0 {
		// 视频不存在，video表中创建这条视频数据
		uploadTime := time.Unix(response.Data.View.Ctime, 0)
		video = models.Video{
			WebSiteId:  WebSiteId,
			Title:      response.Data.View.Title,
			Uuid:       response.Data.View.Bvid,
			CoverUrl:   response.Data.View.Pic,
			VideoDesc:  response.Data.View.Desc,
			CreateTime: time.Now(),
			UploadTime: &uploadTime,
			Duration:   response.Data.View.Duration,
		}
		models.GormDB.Create(&video)
	}
	// 更新视频信息
	video.View = response.Data.View.Stat.View
	video.Danmaku = response.Data.View.Stat.Danmaku
	video.Reply = response.Data.View.Stat.Reply
	video.Favorite = response.Data.View.Stat.Favorite
	video.Coin = response.Data.View.Stat.Coin
	video.Share = response.Data.View.Stat.Share
	video.NowRank = response.Data.View.Stat.NowRank
	video.HisRank = response.Data.View.Stat.HisRank
	video.Like = response.Data.View.Stat.Like
	video.Dislike = response.Data.View.Stat.Dislike
	video.Evaluation = response.Data.View.Stat.Evaluation
	models.GormDB.Model(&video).Updates(map[string]interface{}{
		"View":       response.Data.View.Stat.View,
		"Danmaku":    response.Data.View.Stat.Danmaku,
		"Reply":      response.Data.View.Stat.Reply,
		"Favorite":   response.Data.View.Stat.Favorite,
		"Coin":       response.Data.View.Stat.Coin,
		"Share":      response.Data.View.Stat.Share,
		"NowRank":    response.Data.View.Stat.NowRank,
		"HisRank":    response.Data.View.Stat.HisRank,
		"Like":       response.Data.View.Stat.Like,
		"Dislike":    response.Data.View.Stat.Dislike,
		"Evaluation": response.Data.View.Stat.Evaluation,
	})
	// 更新作者和协作者信息
	if len(response.Data.View.Staff) > 0 {
		DatabaseAuthorInfo := []models.Author{}
		authorIdList := []int64{}
		for _, a := range video.Authors {
			authorIdList = append(authorIdList, a.AuthorId)
		}
		models.GormDB.Where("id in ?", authorIdList).Find(&DatabaseAuthorInfo)
		// models.VideoAuthor 和 response.Data.View.Staff 两边信息做对比，models.VideoAuthor缺少的就添加，models.Author缺少的就添加
		authorHave := false
		for _, b := range response.Data.View.Staff {
			for _, a := range DatabaseAuthorInfo {
				if a.AuthorWebUid == strconv.Itoa(b.Mid) {
					authorHave = true
					break
				}
			}
			if !authorHave {
				// 查询这个作者在Author表中是否存在
				author := models.Author{}
				models.GormDB.Where("author_web_uid = ?", strconv.Itoa(b.Mid)).Find(&author)
				if author.Id == 0 {
					// 作者不存在，数据库中添加作者信息
					author = models.Author{
						AuthorName:   b.Name,
						WebSiteId:    WebSiteId,
						AuthorWebUid: strconv.Itoa(b.Mid),
						Avatar:       b.Face,
						FollowNumber: b.Follower,
					}
					models.GormDB.Create(&author)
				}
				va := models.VideoAuthor{
					Uuid:       response.Data.View.Bvid,
					VideoId:    video.Id,
					AuthorId:   author.Id,
					Contribute: b.Title,
				}
				models.GormDB.Create(&va)
			}
		}
	} else {
		// 没有协作者
		author := response.Data.Card.Card
		AuthorInfo := models.Author{}
		models.GormDB.Where("author_web_uid=?", author.Mid).Find(&AuthorInfo)
		if AuthorInfo.Id == 0 {
			// 作者不存在，数据库中添加作者信息
			AuthorInfo = models.Author{
				AuthorName:   author.Name,
				WebSiteId:    WebSiteId,
				AuthorWebUid: author.Mid,
				Avatar:       author.Face,
				FollowNumber: author.Fans,
				AuthorDesc:   author.Sign,
			}
			models.GormDB.Create(&AuthorInfo)
		}
		if len(video.Authors) > 0 {
			if video.Authors[0].AuthorId != AuthorInfo.Id {
				// 协作者发生变化
				models.GormDB.Model(&video).Association("Authors").Replace(&AuthorInfo)
			}
		} else {
			models.GormDB.Create(&models.VideoAuthor{
				Uuid:       response.Data.View.Bvid,
				VideoId:    video.Id,
				AuthorId:   AuthorInfo.Id,
				Contribute: "UP主",
			})
		}

	}
	// 更新视频标签信息
	var tagHave bool
	for _, v := range response.Data.Tags {
		// 循环video.Tag，如果有这个标签，就标记已存在,并且在videoTag中删除这个标签
		tagHave = false
		for index, tag := range video.Tag {
			if tag.Id == v.TagId {
				tagHave = true
				video.Tag = append(video.Tag[:index], video.Tag[index+1:]...)
				break
			}
		}

		if !tagHave {
			tag := models.Tag{}
			models.GormDB.Find(&tag, "id=?", v.TagId)
			if tag.Name == "" {
				tag.Id = v.TagId
				tag.Name = v.TagName
				models.GormDB.Create(&tag)
			}
			videoTag := models.VideoTag{
				VideoId: video.Id,
				TagId:   v.TagId,
			}
			models.GormDB.Create(&videoTag)
		}

	}

}

func saveBilibiliAuthorVideoList(response bilibili.VideoListPageResponse, WebSiteId, authorId int64, authorVideoUUIDMap map[string]struct{}) {
	if len(response.Data.List.Vlist) == 0 {
		return
	}
	if authorId == 0 {
		authorMid := response.Data.List.Vlist[0].Mid
		// 查询这个作者的id
		models.GormDB.Table("author").
			Select("id").
			Where("author_web_uid = ?", authorMid).
			Find(&authorId)
	}
	if len(authorVideoUUIDMap) == 0 {
		var authorVideoUUIDList []string
		// 查询这个作者本地保存的视频信息
		models.GormDB.Table("video v").
			Select("v.uuid").
			Where("a.author_id = ?", authorId).
			Find(&authorVideoUUIDList)
		for _, videoUuid := range authorVideoUUIDList {
			authorVideoUUIDMap[videoUuid] = member
		}
	}
	var (
		ok          bool
		insertVideo []*models.Video
	)
	for _, videoInfo := range response.Data.List.Vlist {
		_, ok = authorVideoUUIDMap[videoInfo.Bvid]
		if !ok {
			createdTime := time.Unix(videoInfo.Created, 0)
			// 保存视频信息
			vv := models.Video{
				WebSiteId: WebSiteId,
				Authors: []models.VideoAuthor{
					{AuthorId: authorId, Uuid: videoInfo.Bvid},
				},
				Title:      videoInfo.Title,
				VideoDesc:  videoInfo.Description,
				Duration:   bilibili.HourAndMinutesAndSecondsToSeconds(videoInfo.Length),
				Uuid:       videoInfo.Bvid,
				Url:        "",
				CoverUrl:   videoInfo.Pic,
				UploadTime: &createdTime,
				CreateTime: time.Now(),
			}
			insertVideo = append(insertVideo, &vv)
			authorVideoUUIDMap[videoInfo.Bvid] = member
		}
	}
	models.GormDB.Create(insertVideo)
}

func importFileData(fileName string) {
	// 文件是 cmd/spiderProxy/main.go这个tarFolderFile函数打包出来的文件，文件名格式{taskType}_{taskId}.tar.gz
	// 内部包含三种文件 requestParams请求参数 errRequestParams出现错误的请求参数  {taskId}.json结果集，100M一个文件
	defer moveFile(fileName)
	// 解析文件名，获取taskType和taskId
	fileNameList := strings.Split(fileName, "_")
	if len(fileNameList) != 2 {
		utils.ErrorLog.Printf("文件名格式错误：%s\n", fileName)
		return
	}
	taskType := fileNameList[0]
	var aa taskWorker
	switch taskType {
	case baseStruct.VideoDetail:
		aa = &biliVideoDetail{}
		aa.initStruct(taskType)
		gzFileUnzip(path.Join(importingPath, fileName), fileNameList[1], aa)
	case baseStruct.AuthorVideoList:
		aa = &biliAuthorVideoList{}
		aa.initStruct(taskType)
		gzFileUnzip(path.Join(importingPath, fileName), fileNameList[1], aa)
	}
	aa.endOffWorker()

}

// tar.gz文件解压
func gzFileUnzip(fileNamePath, taskId string, handler taskWorker) {
	defer handler.endOffWorker()
	tarFile, err := os.Open(fileNamePath)
	if err != nil {
		utils.ErrorLog.Printf("打开文件失败：%s\n", err.Error())
		return
	}
	defer tarFile.Close()
	gzRead, err := gzip.NewReader(tarFile)
	if err != nil {
		utils.ErrorLog.Printf("解压文件失败：%s\n", err.Error())
		return
	}
	defer gzRead.Close()
	tarRead := tar.NewReader(gzRead)

	for {
		hdr, err := tarRead.Next()
		switch {
		case err == io.EOF:
			return
		case err != nil:
			return
		case hdr == nil:
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir: // 如果是目录时候，跳过
			continue
		case tar.TypeReg: // 如果是文件就结果按行输出到外部
			switch {
			case hdr.Name == "requestParams":
				handler.requestHandle([]byte{})
			case hdr.Name == "errRequestParams":
				handler.errorRequestHandle([]byte{})
			case strings.HasSuffix(hdr.Name, "json"):
				bufioRead := newReaderJSONFile(tarRead)
				for {
					byteData, _, err := bufioRead.line()
					if err == io.EOF {
						break
					}
					if err != nil {
						utils.ErrorLog.Printf("读取文件行失败：%s\n", err.Error())
						continue
					}
					handler.responseHandle(byteData)
				}
			}
		}

	}

}

// 将这个文件移动到finishImportPath目录下
func moveFile(fileName string) {
	err := os.Rename(path.Join(importingPath, fileName), path.Join(finishImportPath, fileName))
	if err != nil {
		utils.ErrorLog.Printf("移动文件失败：%s\n", err.Error())
	}
}

type taskWorker interface {
	initStruct(string)
	requestHandle([]byte)
	errorRequestHandle([]byte)
	responseHandle([]byte)
	endOffWorker()
}

type biliAuthorVideoList struct {
	notRequestParams    []string
	apiType             string
	webSiteId           int64
	authorId            int64
	authorVideoUUIDList map[string]struct{}
}

func (avl *biliAuthorVideoList) initStruct(apiType string) {
	avl.notRequestParams = []string{}
	avl.apiType = apiType
	models.GormDB.Table("web_site").Select("id").Where("web_name = ?", "bilibili").Find(&avl.webSiteId)
}
func (avl *biliAuthorVideoList) requestHandle(data []byte) {
	avl.notRequestParams = append(avl.notRequestParams, string(data))
}
func (avl *biliAuthorVideoList) errorRequestHandle(data []byte) {
	avl.notRequestParams = append(avl.notRequestParams, string(data))

}
func (avl *biliAuthorVideoList) responseHandle(data []byte) {
	response := bilibili.VideoListPageResponse{}
	err := json.Unmarshal(data, &response)
	if err != nil {
		utils.ErrorLog.Printf("解析响应失败：%s\n", err.Error())
		return
	}
	if response.Code != 0 {
		return
	}
	if len(response.Data.List.Vlist) == 0 {
		return
	}
	if avl.authorId == 0 {
		models.GormDB.Model(&models.Author{}).Where("author_web_uid = ?", response.Data.List.Vlist[0].Mid).Find(&avl.authorId)
	}
	saveBilibiliAuthorVideoList(response, avl.webSiteId, avl.authorId, avl.authorVideoUUIDList)

}
func (avl *biliAuthorVideoList) endOffWorker() {

}

type biliVideoDetail struct {
	notRequestParams []string
	apiType          string
	webSiteId        int64
}

func (vd *biliVideoDetail) initStruct(apiType string) {
	vd.notRequestParams = []string{}
	vd.apiType = apiType
	models.GormDB.Table("web_site").Select("id").Where("web_name = ?", "bilibili").Find(&vd.webSiteId)
}
func (vd *biliVideoDetail) requestHandle(data []byte) {
	vd.notRequestParams = append(vd.notRequestParams, string(data))
}
func (vd *biliVideoDetail) errorRequestHandle(data []byte) {
	vd.notRequestParams = append(vd.notRequestParams, string(data))
}
func (vd *biliVideoDetail) responseHandle(data []byte) {
	response := struct {
		Url      string
		Response bilibili.VideoDetailResponse
	}{}
	err := json.Unmarshal(data, &response)
	if err != nil {
		utils.ErrorLog.Printf("解析响应失败：%s\n", err.Error())
		return
	}
	if response.Response.Code != 0 {
		return
	}
	updateBilibiliVideoDetailInfo(response.Response, vd.webSiteId)

}
func (vd *biliVideoDetail) endOffWorker() {

}

func newReaderJSONFile(rd io.Reader) readJsonFile {
	rf := readJsonFile{}
	rf.readObject = rd
	return rf
}

type readJsonFile struct {
	readObject io.Reader
	cache      []byte
}

// 读取一个完整的json对象，从123->{字符读到125->}字符。两边的字符必须是对称出现，返回这样的一个json字符串
func (ro readJsonFile) line() ([]byte, int, error) {
	var buf bytes.Buffer
	started := false
	count := 0

	for {
		b := make([]byte, 1)
		_, err := ro.readObject.Read(b)
		if err != nil {
			if err == io.EOF {
				return buf.Bytes(), buf.Len(), io.EOF
			}
			return nil, 0, err
		}

		if b[0] == '{' {
			started = true
			count++
		}

		if started {
			buf.Write(b)
		}

		if b[0] == '}' {
			count--
			if count == 0 {
				break
			}
		}
	}

	return buf.Bytes(), buf.Len(), nil
}
