package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

var (
	config *utils.Config
)

func readConfig() error {
	fileData, err := os.ReadFile("E:\\GoCode\\videoDynamicAcquisition\\cmd\\ImportProxyData\\config.json")
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
	models.OpenRedis()
}

func main() {
	//for {
	//	readPath()
	//	time.Sleep(time.Minute * 30)
	//}
	waitImportPath = path.Join(config.ProxyDataRootPath, utils.WaitImportPrefix)
	importingPath = path.Join(config.ProxyDataRootPath, utils.ImportingPrefix)
	finishImportPath = path.Join(config.ProxyDataRootPath, utils.FinishImportPrefix)
	errorImportPrefix = path.Join(config.ProxyDataRootPath, utils.ErrorImportPrefix)
	waitImportFileList, err := os.ReadDir(importingPath)
	if err != nil {
		println(err.Error())
		return
	}
	println(time.Now().Format("2006.01.02 15:04:05"))
	errorRequestSaveFile = &utils.WriteFile{
		FolderPrefix:   []string{config.ProxyDataRootPath},
		FileNamePrefix: "errorRequestParams",
	}
	for _, waitImportFile := range waitImportFileList {
		// 开始导入数据
		importFileData(waitImportFile.Name())
	}
	println(time.Now().Format("2006.01.02 15:04:05"))
	//for w := 1; w <= 10; w++ {
	//	go tenWorker()
	//}
	//for {
	//	if len(fileNameChan) == 0 {
	//		break
	//	}
	//	fmt.Printf("当前剩余%d个任务", len(fileNameChan))
	//	time.Sleep(time.Minute * 30)
	//}
}

var fileNameChan chan string

func tenWorker() {
	for {
		select {
		case fileName := <-fileNameChan:
			importFileData(fileName)
		}
	}
}
