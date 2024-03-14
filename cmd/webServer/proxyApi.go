package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"io/ioutil"
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

// {"url":"","response":}
var (
	prefixByte   = []byte{123, 34, 117, 114, 108, 34, 58, 34}
	bracketsByte = []byte{125}
	suffixByte   = []byte{34, 44, 34, 114, 101, 115, 112, 111, 110, 115, 101, 34, 58}
	config       *utils.Config
)

func writeRequestUrl(url string, responseBody []byte) []byte {
	var result []byte
	if len(responseBody) == 0 || responseBody == nil {
		responseBody = []byte{123, 125}
	}
	result = append(result, prefixByte...)
	result = append(result, url...)
	result = append(result, suffixByte...)
	result = append(result, responseBody...)
	result = append(result, bracketsByte...)
	result = append(result, 10)
	return result

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
	folderName := baseStruct.AuthorVideoList
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

	go getAuthorVideoList(requestBody.IdList, folderName, taskId)
	ctx.JSONP(200, map[string]string{"taskId": taskId})
}

func getVideoDetailApi(ctx *gin.Context) {
	requestBody := IdListRequest{}
	err := ctx.ShouldBind(&requestBody)
	if err != nil {
		ctx.JSONP(403, map[string]string{"msg": "获取请求参数失败"})
		return
	}
	folderName := baseStruct.VideoDetail
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
	go getVideoDetailList(requestBody.IdList, folderName, taskId)
	ctx.JSONP(200, map[string]string{"taskId": taskId})

}

func getAuthorVideoList(videoUid []string, folderName, taskId string) {
	defer func() {
		if err := recover(); err != nil {
			utils.ErrorLog.Printf("getVideoDetailList panic %s", err)
			os.WriteFile(path.Join(baseStruct.RootPath, folderName, taskId, "funcError"), []byte(err.(error).Error()), fs.ModePerm)
		}
	}()
	file := utils.WriteFile{
		FolderPrefix:   []string{baseStruct.RootPath, folderName, taskId},
		FileNamePrefix: folderName,
	}
	defer file.Close()
	utils.Info.Println("开始运行")
	// 请求出错的id写入到文件中，文件名 errRequestParams
	errRequestParams, _ := os.Open(path.Join(baseStruct.RootPath, folderName, taskId, "errRequestParams"))
	defer errRequestParams.Close()
	requestParamsFileName := path.Join(baseStruct.RootPath, folderName, taskId, "requestParams")
	os.WriteFile(requestParamsFileName, []byte(strings.Join(videoUid, "\n")), fs.ModePerm)
	for _, i := range videoUid {
		pageIndex := 1
		maxPage := 1
		for pageIndex <= maxPage {
			response, err, requestUrl := bilibili.GetAuthorAllVideoListByByte(i, pageIndex)
			pageIndex++
			file.Write(writeRequestUrl(requestUrl, response))
			if err != nil {
				utils.ErrorLog.Println(err.Error())
				errRequestParams.Write([]byte(i + "\n"))
				pageIndex = maxPage + 1
				continue
			}
			if maxPage == 1 {
				responseBody := new(bilibili.VideoListPageResponse)
				err = json.Unmarshal(response, responseBody)
				if err != nil {
					utils.ErrorLog.Println("解析响应失败")
					utils.ErrorLog.Println(err.Error())
					continue
				}
				if responseBody.Code != 0 {
					continue
				}
				if responseBody.Data.Page.Pn > maxPage {
					maxPage = responseBody.Data.Page.Pn
				}
			}

			time.Sleep(time.Second * 3)
		}
		utils.Info.Printf("%s 爬取完成", i)
	}
	file.Close()
	tarFolderFile(folderName, taskId)
}

func getVideoDetailList(videoUid []string, folderName, taskId string) {
	defer func() {
		if err := recover(); err != nil {
			utils.ErrorLog.Printf("getVideoDetailList panic %s", err)
		}
	}()
	file := utils.WriteFile{
		FolderPrefix:   []string{baseStruct.RootPath, folderName, taskId},
		FileNamePrefix: folderName,
	}
	defer file.Close()
	utils.Info.Println("开始爬取视频详情")
	// 请求出错的id写入到文件中，文件名 errRequestParams
	errRequestParams, _ := os.Open(path.Join(baseStruct.RootPath, folderName, taskId, "errRequestParams"))
	defer errRequestParams.Close()
	fileName := path.Join(baseStruct.RootPath, folderName, taskId, "requestParams")
	slice2 := make([]string, len(videoUid))
	copy(slice2, videoUid)
	os.WriteFile(fileName, []byte(strings.Join(slice2, "\n")), fs.ModePerm)
	for _, i := range videoUid {
		data, requestUrl := bilibili.GetVideoDetailByByte(i)
		file.Write(writeRequestUrl(requestUrl, data))
		if data == nil {
			errRequestParams.Write([]byte(i))
			errRequestParams.Write([]byte{10})
		}
		time.Sleep(time.Second * 4)
		utils.Info.Printf("%s 爬取完成", i)
	}
	file.Close()
	tarFolderFile(folderName, taskId)
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
	writeFilePath := path.Join(baseStruct.RootPath, taskType, taskId)
	fileName := path.Join(config.ProxyDataRootPath, taskType, fmt.Sprintf("%s_%s.tar.gz", taskType, taskId))
	var (
		err error
	)
	if _, err = os.Stat(writeFilePath); os.IsNotExist(err) {
		ctx.JSONP(200, map[string]interface{}{"status": -1, "msg": "任务不存在"})
		return
	}
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		ctx.JSONP(200, map[string]int{"status": 0})
		return
	}
	// 获取打包后的文件MD5
	// 获取文件MD5
	md5, err := utils.GetFileMd5(fileName)
	if err != nil {
		println(err.Error())
		ctx.JSONP(200, map[string]interface{}{"status": 1, "msg": "获取文件MD5失败"})
		return
	}

	ctx.JSONP(200, map[string]interface{}{"status": 1, "md5": md5})
	return
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
	targetFile := path.Join(config.ProxyDataRootPath, folderName, fmt.Sprintf("%s_%s.tar.gz", folderName, taskId))
	baseFolder := path.Join(baseStruct.RootPath, folderName, taskId)
	// 创建目标文件
	file, err := os.Create(targetFile)
	if err != nil {
		utils.ErrorLog.Println(err.Error())
		return
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
		utils.ErrorLog.Println(err.Error())
		return
	}
	perm := os.FileMode(0644) // 设置文件权限为 644 其他用户只有读权限
	os.Chmod(targetFile, perm)
}

func deleteFile() {
	// 每天中午十二点执行
	now := time.Now()
	// 生成明天的中午时间点
	tomorrow := now.AddDate(0, 0, 1)
	noonTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, tomorrow.Location())
	// 计算时间差
	diff := noonTime.Sub(now)
	time.Sleep(diff)
	deleteAfterDayFile(config.ProxyDataRootPath)
	deleteAfterDayFile(path.Join(baseStruct.RootPath, baseStruct.VideoDetail))
	deleteAfterDayFile(path.Join(baseStruct.RootPath, baseStruct.AuthorVideoList))
	deleteFile()
}

func deleteAfterDayFile(folderPath string) {
	// 获取文件夹中的文件列表
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		utils.ErrorLog.Println(err.Error())
		return
	}
	// 遍历文件列表
	for _, file := range files {
		// 获取文件名
		fileName := file.Name()
		// 获取文件创建时间
		fileModifyTime := file.ModTime()
		// 获取当前时间
		currentTime := time.Now()
		// 计算文件创建时间和当前时间的差值
		diff := currentTime.Sub(fileModifyTime)
		// 如果差值大于一天，则删除文件
		if diff.Hours() >= 24 {
			fmt.Printf("%s文件上次修改时间是%s，准备删除\n", fileName, fileModifyTime.Format("2006-01-02-15-04-05"))
			//err := os.Remove(path.Join(folderPath, fileName))
			//if err != nil {
			//	utils.ErrorLog.Println(err.Error())
			//}
		}
	}
}

var fileIndex int = 1

func bilibiliRecommendVideoSave(ctx *gin.Context) {
	// {"ip":requestIp,"time":timeNow,"data":requestBody}
	// 获取推荐视频
	requestBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSONP(501, map[string]interface{}{"code": -1})
		return
	}
	requestIp := ctx.ClientIP()
	timeNow := time.Now()
	file := utils.WriteFile{
		FolderPrefix:   []string{baseStruct.RootPath, baseStruct.RecommendVideo},
		FileNamePrefix: "RecommendVideo",
		FileName: func(lastFileName string) string {
			if lastFileName == "" {
				return "RecommendVideo"
			}
			fileIndex++
			return fmt.Sprintf("RecommendVideo%d", fileIndex)
		},
	}
	writeData := []byte(fmt.Sprintf(`{"ip":"%s","time":%d,"data":"%s"}`, requestIp, timeNow.Unix(), string(requestBody)))
	file.WriteLine(writeData)
	file.Close()
	ctx.JSONP(200, map[string]interface{}{"code": 0})
	return
}
