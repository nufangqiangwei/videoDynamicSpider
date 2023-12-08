package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/bilibili"
	"videoDynamicAcquisition/utils"
)

//const rootPath = "C:\\Code\\GO\\videoDynamicSpider\\cmd\\appendBiliVideo"
const maxFileSize = 100 * 1024 * 1024

// {"url":"","response":""}
//var (
//	prefixByte   = []byte{123, 34, 117, 114, 108, 34, 58, 34}
//	bracketsByte = []byte{125}
//	suffixByte   = []byte{34, 44, 34, 114, 101, 115, 112, 111, 110, 115, 101, 34, 58}
//)

type WriteFile struct {
	folderPrefix   []string
	fileNamePrefix string
	file           *os.File
}

func (wf *WriteFile) checkFileSize() {
	if wf.file == nil {
		filePath := append(wf.folderPrefix, fmt.Sprintf("%s-%s.json", wf.fileNamePrefix, time.Now().Format("2006-01-02-15-04-05")))
		f, err := os.OpenFile(path.Join(filePath...), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			utils.ErrorLog.Printf("打开新文件失败%s", err.Error())
			panic(err)
		}
		wf.file = f
		return
	}
	for {
		fi, err := wf.file.Stat()
		if err != nil {
			utils.ErrorLog.Printf("获取文件信息失败%s", err.Error())
			panic(err)
		}
		if fi.Size() >= maxFileSize {
			wf.file.Close()
			filePath := append(wf.folderPrefix, fmt.Sprintf("%s-%s.json", wf.fileNamePrefix, time.Now().Format("2006-01-02-15-04-05")))
			f, err := os.OpenFile(path.Join(filePath...), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				utils.ErrorLog.Printf("打开新文件失败%s", err.Error())
				panic(err)
			}
			wf.file = f
		} else {
			break
		}
	}
}

func main() {
	utils.InitLog(baseStruct.RootPath)
	server := gin.Default()
	server.POST("getAuthorAllVideo", getAuthorAllVideo)
	server.POST("getVideoDetail", getVideoDetailApi)
	server.GET("getTaskStatus", getTaskStatus)
	server.Run(":9000")
}

type IdListRequest struct {
	IdList []string
}

func getAuthorAllVideo(ctx *gin.Context) {
	requestBody := IdListRequest{}
	err := ctx.ShouldBind(&requestBody)
	if err != nil {
		ctx.JSONP(403, map[string]string{"msg": "获取请求参数失败", "taskId": ""})
		return
	}
	folderName := "allVideo"
	createFolder(false, folderName)

	taskId := uuid.NewString()
	err = createFolder(true, folderName, taskId)
	for err != nil {
		if err.Error() == "文件夹已存在" {
			taskId = uuid.NewString()
			err = createFolder(true, folderName, taskId)
		} else {
			ctx.JSONP(503, map[string]string{"msg": "创建文件夹失败", "taskId": ""})
			return
		}
	}

	videoChan := make(chan []byte, 5)

	go func() {
		defer close(videoChan)
		// 请求出错的id写入到文件中，文件名 errRequestParams
		errRequestParams, _ := os.Open(path.Join(baseStruct.RootPath, folderName, taskId, "errRequestParams"))
		defer errRequestParams.Close()
		fileName := path.Join(baseStruct.RootPath, folderName, taskId, "requestParams")
		slice2 := make([]string, len(requestBody.IdList))
		copy(slice2, requestBody.IdList)
		os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
		for _, i := range requestBody.IdList {
			err := bilibili.GetAuthorAllVideoListTOJSON(i, videoChan)
			if err != nil {
				errRequestParams.Write([]byte(i))
				errRequestParams.Write([]byte{10})
				utils.ErrorLog.Printf("爬取失败%s", err.Error())
			}
			time.Sleep(time.Second * 10)
			utils.Info.Printf("%s 爬取完成", i)
			if len(slice2) > 0 {
				slice2 = slice2[1:]
				os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
			}
		}
	}()
	go func() {
		file := WriteFile{
			folderPrefix:   []string{baseStruct.RootPath, folderName, taskId},
			fileNamePrefix: folderName,
		}
		file.checkFileSize()
		defer file.file.Close()
		utils.Info.Println("开始运行")
		for i := range videoChan {
			if i == nil {
				file.checkFileSize()
				continue
			}
			file.file.Write(i)
			file.file.Write([]byte{10})
		}
		file.file.Close()
		tarFolderFile(folderName, taskId)
	}()

	ctx.JSONP(200, map[string]string{"taskId": taskId})
}

func getVideoDetailApi(ctx *gin.Context) {
	requestBody := IdListRequest{}
	err := ctx.ShouldBind(&requestBody)
	if err != nil {
		ctx.JSONP(403, map[string]string{"msg": "获取请求参数失败"})
		return
	}
	folderName := "videoDetail"
	createFolder(false, folderName)
	taskId := uuid.NewString()
	err = createFolder(true, folderName, taskId)
	for err != nil {
		if err.Error() == "文件夹已存在" {
			taskId = uuid.NewString()
			err = createFolder(true, folderName, taskId)
		} else {
			ctx.JSONP(503, map[string]string{"msg": "创建文件夹失败"})
			return
		}
	}
	go func() {
		getVideoDetailList(requestBody.IdList, folderName, taskId)
		tarFolderFile(folderName, taskId)
	}()
	ctx.JSONP(200, map[string]string{"taskId": taskId})

}

func getVideoDetailList(videoUid []string, folderName, taskId string) {
	resultChan := make(chan []byte, 5)
	go func() {
		defer close(resultChan)
		// 请求出错的id写入到文件中，文件名 errRequestParams
		errRequestParams, _ := os.Open(path.Join(baseStruct.RootPath, folderName, taskId, "errRequestParams"))
		defer errRequestParams.Close()
		fileName := path.Join(baseStruct.RootPath, folderName, taskId, "requestParams")
		slice2 := make([]string, len(videoUid))
		copy(slice2, videoUid)
		os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
		for _, i := range videoUid {
			data := bilibili.GetVideoDetailByByte(i)
			if data == nil {
				errRequestParams.Write([]byte(i))
				errRequestParams.Write([]byte{10})
				continue
			}
			resultChan <- data
			time.Sleep(time.Second * 4)
			utils.Info.Printf("%s 爬取完成", i)
			if len(slice2) > 0 {
				slice2 = slice2[1:]
				os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
			}
		}
	}()
	file := WriteFile{
		folderPrefix:   []string{baseStruct.RootPath, folderName, taskId},
		fileNamePrefix: folderName,
	}
	file.checkFileSize()
	defer file.file.Close()
	utils.Info.Println("开始爬取视频详情")
	for i := range resultChan {
		if i == nil {
			file.checkFileSize()
			continue
		}
		file.file.Write(i)
		file.file.Write([]byte{10})
	}
}

func getTaskStatus(ctx *gin.Context) {
	taskId := ctx.Query("taskId")
	taskType := ctx.Query("taskType")
	if taskType == "" {
		ctx.JSONP(200, map[string]interface{}{"status": -1, "msg": "任务类型不存在"})
		return
	}
	if taskId == "" {
		ctx.JSONP(200, map[string]interface{}{"status": -1, "msg": "任务id不存在"})
		return
	}
	fileName := path.Join(baseStruct.RootPath, taskType, taskId, "requestParams")
	var (
		file os.FileInfo
		err  error
	)
	if file, err = os.Stat(fileName); os.IsNotExist(err) {
		ctx.JSONP(200, map[string]interface{}{"status": -1, "msg": "任务不存在"})
		return
	}

	if file.Size() == 0 {
		// 获取打包后的文件MD5
		fileName = path.Join(baseStruct.RootPath, taskType, fmt.Sprintf("%s.tar.gz", taskId))
		// 获取文件MD5
		md5, err := utils.GetFileMd5(fileName)
		if err != nil {
			ctx.JSONP(200, map[string]interface{}{"status": 1, "msg": "获取文件MD5失败"})
			return
		}

		ctx.JSONP(200, map[string]interface{}{"status": 1, "md5": md5})
		return
	}
	ctx.JSONP(200, map[string]int{"status": 0})
}

func createFolder(haveReturnError bool, elem ...string) error {
	elem = append([]string{baseStruct.RootPath}, elem...)
	dir := path.Join(elem...)
	// 判断文件夹是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 文件夹不存在，创建文件夹
		err := os.Mkdir(dir, 0755)
		if err != nil {
			fmt.Println("创建文件夹失败:", err)
			return err
		}
		fmt.Println("文件夹创建成功")
	} else {
		if haveReturnError {
			return errors.New("文件夹已存在")
		}
		fmt.Println("文件夹已存在")
	}
	return nil
}

func tarFolderFile(folderName, taskId string) {
	// 设置源文件夹和目标文件名
	sourceFolder := path.Join(baseStruct.RootPath, folderName, taskId)
	targetFile := path.Join(baseStruct.RootPath, folderName, fmt.Sprintf("%s|%s.tar.gz", folderName, taskId))
	baseFolder := path.Join(baseStruct.RootPath, folderName, taskId)
	// 创建目标文件
	file, err := os.Create(targetFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 创建gzip写入器
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// 创建tar写入器
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// 遍历源文件夹
	err = filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 获取文件相对于指定文件夹的路径
		relPath, err := filepath.Rel(baseFolder, path)
		if err != nil {
			return err
		}

		// 获取文件头信息
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}

		// 写入文件头
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		// 如果不是目录，则写入文件内容
		if !info.IsDir() {
			// 打开源文件
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// 将源文件内容复制到tar文件中
			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}
