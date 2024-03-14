package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/panjf2000/ants/v2"
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
	member               void
	waitImportPath       string
	importingPath        string
	finishImportPath     string
	errorImportPrefix    string
	errorRequestSaveFile *utils.WriteFile
	limitGoroutine       *ants.Pool
	task                 *taskRequestWorker
)

func readPath() {
	// 读取 waitImportFile这个目录下的文件，有新的文件出现就将这个文件移动到importingFile目录下，然后开始导入数据
	// 导入完成后，将这个文件移动到finishImportFile目录下
	waitImportPath = path.Join(config.ProxyDataRootPath, utils.WaitImportPrefix)
	importingPath = path.Join(config.ProxyDataRootPath, utils.ImportingPrefix)
	finishImportPath = path.Join(config.ProxyDataRootPath, utils.FinishImportPrefix)
	errorImportPrefix = path.Join(config.ProxyDataRootPath, utils.ErrorImportPrefix)
	if errorRequestSaveFile == nil {
		errorRequestSaveFile = &utils.WriteFile{
			FolderPrefix:   []string{errorImportPrefix},
			FileNamePrefix: "errorRequestParams",
		}
	}
	if limitGoroutine == nil {
		limitGoroutine, _ = ants.NewPool(10)
	}
	if task != nil {
		task = taskRequestWorkerNew()
	}
	// 检查importingPath目录下的文件是否有上次异常退出残留下来的文件
	importingFileList, err := os.ReadDir(importingPath)
	if err != nil {
		utils.ErrorLog.Printf("读取目录失败：%s\n", err.Error())
	} else {
		for _, importingFile := range importingFileList {
			if strings.HasSuffix(importingFile.Name(), "tar.gz") {
				time.Sleep(time.Microsecond * 500)
				fileName := importingFile.Name()
				importFileData(fileName)
				//err = limitGoroutine.Submit(func() {
				//	importFileData(fileName)
				//})
				//if err != nil {
				//	utils.ErrorLog.Printf("协程池添加失败：%s，文件名%s\n", err.Error(), importingFile.Name())
				//}
			}
		}
	}
	// 读取waitImportPath目录下的文件
	waitImportFileList, err := os.ReadDir(waitImportPath)
	if err != nil {
		utils.ErrorLog.Printf("读取目录失败：%s\n", err.Error())
		waitImportFileList = make([]os.DirEntry, 0)
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
				continue
			}
			time.Sleep(time.Microsecond * 500)
			fileName := waitImportFile.Name()
			// 开始导入数据
			importFileData(fileName)
			//err = limitGoroutine.Submit(func() {
			//	importFileData(fileName)
			//})
			if err != nil {
				utils.ErrorLog.Printf("协程池添加失败：%s，文件名%s\n", err.Error(), waitImportFile.Name())
			}
		}
	}
	limitGoroutine.Running()

}

func importFileData(fileName string) {
	// 文件是 cmd/webServer/main.go这个tarFolderFile函数打包出来的文件，文件名格式{taskType}_{taskId}.tar.gz
	// 内部包含三种文件 requestParams请求参数 errRequestParams出现错误的请求参数  {taskId}.json结果集，100M一个文件
	utils.Info.Println("importFileData函数开始解析", fileName)
	// 解析文件名，获取taskType和taskId
	fileNameList := strings.Split(fileName, "_")
	if len(fileNameList) != 2 {
		utils.ErrorLog.Printf("文件名格式错误：%s\n", fileName)
		moveFile(importingPath, errorImportPrefix, fileName)
		return
	}
	taskType := fileNameList[0]
	var aa taskWorker
	switch taskType {
	case baseStruct.VideoDetail:
		aa = &biliVideoDetail{}
	case baseStruct.AuthorVideoList:
		aa = &biliAuthorVideoList{}
	}
	aa.initStruct(taskType)
	err := gzFileUnzip(path.Join(importingPath, fileName), fileNameList[1], aa)
	if err != nil {
		moveFile(importingPath, errorImportPrefix, fileName)
	} else {
		moveFile(importingPath, finishImportPath, fileName)
	}
	aa.endOffWorker()
	utils.Info.Println("importFileData函数 ", fileName, "解析完成")
}

// tar.gz文件解压
func gzFileUnzip(fileNamePath, taskId string, handler taskWorker) error {
	defer handler.endOffWorker()
	tarFile, err := os.Open(fileNamePath)
	if err != nil {
		utils.ErrorLog.Printf("打开%s文件失败：%s\n", fileNamePath, err.Error())
		return err
	}
	defer tarFile.Close()
	gzRead, err := gzip.NewReader(tarFile)
	if err != nil {
		utils.ErrorLog.Printf("解压%s文件失败：%s\n", fileNamePath, err.Error())
		return err
	}
	defer gzRead.Close()
	tarRead := tar.NewReader(gzRead)
	var jsonDataError error
	for {
		hdr, err := tarRead.Next()
		switch {
		case err == io.EOF:
			return jsonDataError
		case err != nil:
			return nil
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
				bufioRead := utils.NewReaderJSONFile(tarRead)
				for {
					byteData, _, err := bufioRead.Line()
					if err == io.EOF {
						break
					}
					if err != nil {
						utils.ErrorLog.Printf("读取%s文件行失败：%s\n", fileNamePath, err.Error())
						continue
					}

					err = handler.responseHandle(byteData)
					if err != nil {
						jsonDataError = err
						if err.Error() == "json错误" {
							break
						}
					}
				}
			}
		}

	}

}

// 将这个文件移动到finishImportPath目录下
func moveFile(sourcePath, targetPath, fileName string) {
	err := os.Rename(path.Join(sourcePath, fileName), path.Join(targetPath, fileName))
	if err != nil {
		utils.ErrorLog.Printf("移动文件失败：%s\n", err.Error())
	}
}

type taskRequestWorker struct {
	authorVideoListResponseChan chan []byte
	videoDetailResponseChan     chan []byte
}

func taskRequestWorkerNew() *taskRequestWorker {
	t := &taskRequestWorker{
		authorVideoListResponseChan: make(chan []byte, 100),
		videoDetailResponseChan:     make(chan []byte, 100),
	}
	for i := 0; i < 10; i++ {
		go func() {
			a := biliVideoDetail{}
			a.initStruct("videoDetail")
			for {
				select {
				case data := <-t.videoDetailResponseChan:
					err := a.responseHandle(data)
					if err != nil {
						errorRequestSaveFile.Write(data)
					}
				}
			}
		}()
		go func() {
			a := biliAuthorVideoList{}
			a.initStruct("authorVideoList")
			for {
				select {
				case data := <-t.authorVideoListResponseChan:
					err := a.responseHandle(data)
					if err != nil {
						errorRequestSaveFile.Write(data)
					}
				}
			}
		}()
	}
	return t
}
func (t *taskRequestWorker) sendVideoDetail(data []byte) {
	t.videoDetailResponseChan <- data
}
func (t *taskRequestWorker) sendAuthorVideoList(data []byte) {
	t.authorVideoListResponseChan <- data
}

type taskWorker interface {
	initStruct(string)
	requestHandle([]byte)
	errorRequestHandle([]byte)
	responseHandle([]byte) error
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
func (avl *biliAuthorVideoList) responseHandle(data []byte) error {
	response := bilibili.VideoListPageResponse{}
	err := json.Unmarshal(data, &response)
	if err != nil {
		utils.ErrorLog.Printf("解析响应失败：%s\n", err.Error())
		errorRequestSaveFile.WriteLine(data)
		return err
	}
	if response.Code != 0 {
		errorRequestSaveFile.WriteLine(data)
		return errors.New("请求出错")
	}
	if len(response.Data.List.Vlist) == 0 {
		errorRequestSaveFile.WriteLine(data)
		return nil
	}
	if avl.authorId == 0 {
		models.GormDB.Model(&models.Author{}).Where("author_web_uid = ?", response.Data.List.Vlist[0].Mid).Find(&avl.authorId)
	}
	return avl.saveAuthorVideoList(response)
}
func (avl *biliAuthorVideoList) endOffWorker() {

}
func (avl *biliAuthorVideoList) saveAuthorVideoList(response bilibili.VideoListPageResponse) error {
	if len(response.Data.List.Vlist) == 0 {
		return nil
	}
	var err error
	if avl.authorId == 0 {
		authorMid := response.Data.List.Vlist[0].Mid
		// 查询这个作者的id
		err = models.GormDB.Table("author").
			Select("id").
			Where("author_web_uid = ?", authorMid).
			Find(&avl.authorId).Error
		if err != nil {
			utils.ErrorLog.Printf("查询作者id失败：%s\n", err.Error())
			return err
		}
	}
	if len(avl.authorVideoUUIDList) == 0 {
		var authorVideoUUIDList []string
		// 查询这个作者本地保存的视频信息
		err = models.GormDB.Table("video v").
			Select("v.uuid").
			Where("a.author_id = ?", avl.authorId).
			Find(&authorVideoUUIDList).Error
		if err != nil {
			utils.ErrorLog.Printf("查询作者视频信息失败：%s\n", err.Error())
			return err
		}
		for _, videoUuid := range authorVideoUUIDList {
			avl.authorVideoUUIDList[videoUuid] = member
		}
	}
	var (
		ok          bool
		insertVideo []*models.Video
	)
	for _, videoInfo := range response.Data.List.Vlist {
		_, ok = avl.authorVideoUUIDList[videoInfo.Bvid]
		if !ok {
			createdTime := time.Unix(videoInfo.Created, 0)
			// 保存视频信息
			vv := models.Video{
				WebSiteId: avl.webSiteId,
				Authors: []models.VideoAuthor{
					{AuthorId: avl.authorId, Uuid: videoInfo.Bvid},
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
			avl.authorVideoUUIDList[videoInfo.Bvid] = member
		}
	}
	err = models.GormDB.Create(insertVideo).Error
	if err != nil {
		utils.ErrorLog.Printf("保存视频信息失败：%s\n", err.Error())
	}
	return err
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
func (vd *biliVideoDetail) responseHandle(data []byte) error {
	//response := struct {
	//	Url      string
	//	Response bilibili.VideoDetailResponse
	//}{}
	response := bilibili.VideoDetailResponse{}
	err := json.Unmarshal(data, &response)
	if err != nil {
		utils.ErrorLog.Printf("解析响应失败：%s\n", err.Error())
		errorRequestSaveFile.WriteLine(data)
		return errors.New("json错误")
	}
	if response.Code != 0 {
		if response.Code == 62002 || response.Code == -404 {
			return nil
		}
		errorRequestSaveFile.WriteLine(data)
		return nil
	}
	_ = vd.updateVideoDetailInfo(response)
	return nil
}
func (vd *biliVideoDetail) endOffWorker() {

}
func (vd *biliVideoDetail) updateVideoDetailInfo(response bilibili.VideoDetailResponse) error {
	video := models.Video{}
	var (
		err error
	)

	videoId := models.VideoRedis(response.Data.View.Bvid)
	if videoId == 0 {
		// 视频不存在，video表中创建这条视频数据
		uploadTime := time.Unix(response.Data.View.Ctime, 0)
		video = models.Video{
			WebSiteId:  vd.webSiteId,
			Title:      response.Data.View.Title,
			Uuid:       response.Data.View.Bvid,
			CoverUrl:   response.Data.View.Pic,
			VideoDesc:  response.Data.View.Desc,
			CreateTime: time.Now(),
			UploadTime: &uploadTime,
			Duration:   response.Data.View.Duration,
		}
		err = models.GormDB.Create(&video).Error
		if err != nil {
			utils.ErrorLog.Printf("保存视频信息失败：%s\n", err.Error())
			return err
		}
		models.CreateVideoToRedis(video.Uuid, video.Id)
	} else {
		models.GormDB.Where("id = ?", videoId).Preload("Authors").Preload("Tag").
			Limit(1).Find(&video)
	}
	// 更新视频信息
	err = models.GormDB.Create(&models.VideoPlayData{
		VideoId:    video.Id,
		View:       response.Data.View.Stat.View,
		Danmaku:    response.Data.View.Stat.Danmaku,
		Reply:      response.Data.View.Stat.Reply,
		Favorite:   response.Data.View.Stat.Favorite,
		Coin:       response.Data.View.Stat.Coin,
		Share:      response.Data.View.Stat.Share,
		NowRank:    response.Data.View.Stat.NowRank,
		HisRank:    response.Data.View.Stat.HisRank,
		Like:       response.Data.View.Stat.Like,
		Dislike:    response.Data.View.Stat.Dislike,
		Evaluation: response.Data.View.Stat.Evaluation,
		CreateTime: time.Now(),
	}).Error

	if err != nil {
		utils.ErrorLog.Printf("更新视频信息失败：%s\n", err.Error())
		return err
	}
	// 更新作者和协作者信息
	if len(response.Data.View.Staff) > 0 {
		DatabaseAuthorInfo := []models.Author{}
		authorIdList := []int64{}
		for _, a := range video.Authors {
			authorIdList = append(authorIdList, a.AuthorId)
		}
		err = models.GormDB.Where("id in ?", authorIdList).Find(&DatabaseAuthorInfo).Error
		if err != nil {
			utils.ErrorLog.Printf("查询作者信息失败：%s\n", err.Error())
			return err
		}
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
				author.Id = models.AuthorRedis(strconv.Itoa(b.Mid))
				//err = models.GormDB.Where("author_web_uid = ?", strconv.Itoa(b.Mid)).Find(&author).Error
				//if err != nil {
				//	utils.ErrorLog.Printf("查询作者信息失败：%s\n", err.Error())
				//	return err
				//}
				if author.Id == 0 {
					// 作者不存在，数据库中添加作者信息
					author = models.Author{
						AuthorName:   b.Name,
						WebSiteId:    vd.webSiteId,
						AuthorWebUid: strconv.Itoa(b.Mid),
						Avatar:       b.Face,
						FollowNumber: b.Follower,
					}
					err = models.GormDB.Create(&author).Error
					if err != nil {
						utils.ErrorLog.Printf("保存作者信息失败：%s\n", err.Error())
						return err
					}
					models.CreateVideoAuthorToRedis(author.AuthorWebUid, author.Id)
				}
				va := models.VideoAuthor{
					Uuid:       response.Data.View.Bvid,
					VideoId:    video.Id,
					AuthorId:   author.Id,
					Contribute: b.Title,
				}
				err = models.GormDB.Create(&va).Error
				if err != nil {
					utils.ErrorLog.Printf("保存视频作者信息失败：%s\n", err.Error())
					return err

				}
			}
		}
	} else {
		// 没有协作者
		author := response.Data.Card.Card
		AuthorInfo := models.Author{}
		//err = models.GormDB.Where("author_web_uid=?", author.Mid).Find(&AuthorInfo).Error
		//if err != nil {
		//	utils.ErrorLog.Printf("查询作者信息失败：%s\n", err.Error())
		//	return err
		//}
		AuthorInfo.Id = models.AuthorRedis(author.Mid)
		if AuthorInfo.Id == 0 {
			// 作者不存在，数据库中添加作者信息
			AuthorInfo = models.Author{
				AuthorName:   author.Name,
				WebSiteId:    vd.webSiteId,
				AuthorWebUid: author.Mid,
				Avatar:       author.Face,
				FollowNumber: author.Fans,
				AuthorDesc:   author.Sign,
			}
			err = models.GormDB.Create(&AuthorInfo).Error
			if err != nil {
				utils.ErrorLog.Printf("保存作者信息失败：%s\n", err.Error())
				return err
			}
			models.CreateVideoAuthorToRedis(AuthorInfo.AuthorWebUid, AuthorInfo.Id)
		}
		if len(video.Authors) > 0 {
			if video.Authors[0].AuthorId != AuthorInfo.Id {
				// 协作者发生变化
				err = models.GormDB.Model(&video).Association("Authors").Replace(&AuthorInfo)
				if err != nil {
					utils.ErrorLog.Printf("更新视频作者信息失败：%s\n", err.Error())
					return err
				}
			}
		} else {
			err = models.GormDB.Create(&models.VideoAuthor{
				Uuid:       response.Data.View.Bvid,
				VideoId:    video.Id,
				AuthorId:   AuthorInfo.Id,
				Contribute: "UP主",
			}).Error
			if err != nil {
				utils.ErrorLog.Printf("保存视频作者信息失败：%s\n", err.Error())
				return err
			}
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
	return vd.relatedVideo(response)
}
func (vd *biliVideoDetail) relatedVideo(response bilibili.VideoDetailResponse) error {
	for _, videoInfo := range response.Data.Related {
		videoId := models.VideoRedis(videoInfo.Bvid)
		if videoId > 0 {
			continue
		}
		uploadTime := time.Unix(videoInfo.Ctime, 0)
		video := models.Video{
			WebSiteId:  vd.webSiteId,
			Title:      videoInfo.Title,
			Uuid:       videoInfo.Bvid,
			CoverUrl:   videoInfo.Pic,
			VideoDesc:  videoInfo.Desc,
			CreateTime: time.Now(),
			UploadTime: &uploadTime,
			Duration:   videoInfo.Duration,
			Authors: []models.VideoAuthor{
				{AuthorUUID: strconv.Itoa(videoInfo.Owner.Mid), Contribute: "UP主"},
			},
			StructAuthor: []models.Author{
				{
					WebSiteId:    vd.webSiteId,
					AuthorWebUid: strconv.Itoa(videoInfo.Owner.Mid),
					AuthorName:   videoInfo.Owner.Name,
					Avatar:       videoInfo.Owner.Face,
				},
			},
		}
		err := video.UpdateVideo()
		if err != nil {
			return err
		}
		VideoPlayData := models.VideoPlayData{
			VideoId:    video.Id,
			View:       videoInfo.Stat.View,
			Danmaku:    videoInfo.Stat.Danmaku,
			Reply:      videoInfo.Stat.Reply,
			Favorite:   videoInfo.Stat.Favorite,
			Coin:       videoInfo.Stat.Coin,
			Share:      videoInfo.Stat.Share,
			NowRank:    videoInfo.Stat.NowRank,
			HisRank:    videoInfo.Stat.HisRank,
			Like:       videoInfo.Stat.Like,
			Dislike:    videoInfo.Stat.Dislike,
			CreateTime: time.Now(),
		}
		models.GormDB.Create(&VideoPlayData)
		models.CreateVideoToRedis(video.Uuid, video.Id)
	}
	return nil
}
