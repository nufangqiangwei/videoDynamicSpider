package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

func runToDoTask(interface{}) {
	// 查询正在执行任务的代理
	var proxyTasks []models.ProxySpiderTask
	err := models.GormDB.Where("status = ?", 1).Find(&proxyTasks).Error
	if err != nil {
		log.ErrorLog.Println("分配代理任务中，查询当前运作的代理错误:", err)
		return
	}
	// 如果没有可用的代理任务，结束本次循环
	if len(proxyTasks) == len(config.Proxy) {
		return
	}
	// 从config.Proxy中找到空闲的代理,存入到leisureProxy中
	var leisureProxy []utils.ProxyInfo
	for _, proxy := range config.Proxy {
		// 标记是否找到可用的代理
		assigned := false
		for _, proxyTask := range proxyTasks {
			if proxyTask.SpiderIp == proxy.IP {
				assigned = true
				break
			}
		}
		if !assigned {
			leisureProxy = append(leisureProxy, proxy)
		}
	}
	if len(leisureProxy) == 0 {
		return
	}
	// 查询待执行的任务
	var tasks []models.TaskToDoList
	err = models.GormDB.Where("status = ?", 0).Order("task_type , created_at desc").Limit(len(leisureProxy) * 300).Find(&tasks).Error
	if err != nil {
		fmt.Println("Failed to query tasks:", err)
		return
	}
	// 如果没有待执行的任务，结束本次循环
	if len(tasks) == 0 {
		return
	}
	taskGroup := groupTasks(tasks, 300)
	// 分配任务给代理
	for _, waitRunTask := range taskGroup {
		if len(waitRunTask) == 0 {
			continue
		}
		proxy := leisureProxy[len(leisureProxy)-1]
		leisureProxy = leisureProxy[:len(leisureProxy)-1]
		taskUUID, err := sendTasksToProxy(waitRunTask, proxy)
		if err != nil {
			fmt.Println("请求代理出错:", err.Error())
			continue
		}
		// 添加数据到ProxySpiderTask表中
		proxySpiderTask := models.ProxySpiderTask{
			SpiderIp:       proxy.IP,
			TaskType:       waitRunTask[0].TaskType,
			TaskName:       waitRunTask[0].TaskType,
			TaskId:         taskUUID,
			Status:         1,
			StartTimestamp: time.Now(),
		}
		err = models.GormDB.Create(&proxySpiderTask).Error
		if err != nil {
			fmt.Println("添加数据到ProxySpiderTask表中出错:", err.Error())
			continue
		}
		fmt.Printf("分配 %d 个%s任务给代理 %s\n", len(waitRunTask), waitRunTask[0].TaskType, proxy.IP)
		// TaskToDoList 更新已进行
		taskToDoListID := make([]uint, 0)
		for _, task := range waitRunTask {
			taskToDoListID = append(taskToDoListID, task.ID)
		}
		err = models.GormDB.Model(&models.TaskToDoList{}).Where("id in ?", taskToDoListID).Updates(map[string]interface{}{
			"status": 1, "run_task_id": proxySpiderTask.ID,
		}).Error
		if err != nil {
			fmt.Println("TaskToDoList 更新已进行出错:", err.Error())
			continue
		}
	}
}

// 发送任务给代理进行执行
func sendTasksToProxy(tasks []models.TaskToDoList, proxy utils.ProxyInfo) (string, error) {
	if len(tasks) == 0 {
		return "", errors.New("no tasks to send")
	}
	// 构建请求参数
	idList := make([]string, 0)
	for _, task := range tasks {
		idList = append(idList, task.TaskParams)
	}
	requestBody, err := json.Marshal(map[string]interface{}{
		"IdList": idList,
	})
	if err != nil {
		return "", err
	}

	// 获取代理IP
	proxyIp := proxy.IP

	// 发送POST请求给代理
	url := fmt.Sprintf("%s/%s", proxyIp, tasks[0].TaskType)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 处理响应
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	// 解析响应
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析返回的UUID并更新到任务表中
	var responseData map[string]string
	err = json.Unmarshal(responseBody, &responseData)
	if err != nil {
		return "", err
	}
	taskId, ok := responseData["taskId"]
	if !ok {
		return "", errors.New(responseData["msg"])
	}
	return taskId, nil

}

// 将任务按照TaskType进行分组
func groupTasks(tasks []models.TaskToDoList, size int) [][]models.TaskToDoList {
	groups := make([][]models.TaskToDoList, 0)
	taskMap := make(map[string][]models.TaskToDoList)

	// 将任务按照TaskType进行分组
	for _, task := range tasks {
		taskMap[task.TaskType] = append(taskMap[task.TaskType], task)
	}

	// 将分组后的任务按照指定大小进行切片
	for _, group := range taskMap {
		for i := 0; i < len(group); i += size {
			end := i + size
			if end > len(group) {
				end = len(group)
			}
			groups = append(groups, group[i:end])
		}
	}

	return groups
}

type proxyTaskStatusResponse struct {
	Status int
	Msg    string
	Md5    string
}

// 查询当前正在进行的代理任务是否已经完成
func checkProxyTaskStatus(interface{}) {
	var proxyTasks []models.ProxySpiderTask
	err := models.GormDB.Where("status = ?", 1).Find(&proxyTasks).Error
	if err != nil {
		log.ErrorLog.Println("查询当前正在进行的代理任务是否已经完成错误:", err)
		return
	}
	for _, proxyTask := range proxyTasks {
		// 查询代理任务是否已经完成
		url := fmt.Sprintf("%s/getTaskStatus?taskId=%s&taskType=%s", proxyTask.SpiderIp, proxyTask.TaskId, proxyTask.TaskType)
		resp, err := http.Get(url)
		if err != nil {
			log.ErrorLog.Println("查询代理任务是否已经完成错误:", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.ErrorLog.Println("查询代理任务是否已经完成错误:unexpected response status:", resp.StatusCode)
			continue
		}
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.ErrorLog.Println("查询代理任务是否已经完成错误:", err)
			continue
		}
		responseData := proxyTaskStatusResponse{}
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			log.ErrorLog.Println("查询代理任务是否已经完成错误:", err)
			continue
		}
		if responseData.Status == 1 {
			// 代理任务已经完成，更新ProxySpiderTask表中的数据
			err = models.GormDB.Model(&models.ProxySpiderTask{}).Where("id = ?", proxyTask.ID).Updates(map[string]interface{}{
				"status":          2,
				"end_timestamp":   time.Now(),
				"result_file_md5": responseData.Md5,
			}).Error
			if err != nil {
				log.ErrorLog.Println("更新ProxySpiderTask表中的数据错误:", err)
				continue
			}
		}
	}
}

func downloadProxyTaskDataFile(interface{}) {
	var proxyTasks []models.ProxySpiderTask
	models.GormDB.Where("status = ?", 2).Find(&proxyTasks)
	var err error
	for _, proxyTask := range proxyTasks {
		err = downloadTaskDataFile(proxyTask.TaskId, proxyTask.TaskType, proxyTask.SpiderIp)
		if err == nil {
			// 更新ProxySpiderTask表中的数据
			err = models.GormDB.Model(&models.ProxySpiderTask{}).Where("id = ?", proxyTask.ID).Updates(map[string]interface{}{
				"status": 3,
			}).Error
			if err != nil {
				log.ErrorLog.Println("更新ProxySpiderTask表中的数据错误:", err)
				continue
			}
		}

	}
}

// 下载已完成的任务文件
func downloadTaskDataFile(taskId, taskType, ip string) error {
	fileName := fmt.Sprintf("%s_%s.tar.gz", taskType, taskId)
	// 下载任务文件,nginx作为文件服务器
	url := fmt.Sprintf("%s/downloadTaskDataFile/%s/%s", ip, taskType, fileName)
	resp, err := http.Get(url)
	if err != nil {
		log.ErrorLog.Println(url, "下载任务文件错误:", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.ErrorLog.Println(url, "下载任务文件响应码错误:", resp.StatusCode)
		return errors.New("响应码错误")
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorLog.Println(url, "下载任务文件错误:", err)
		return err
	}
	// 将文件保存到本地
	err = ioutil.WriteFile(path.Join(config.ProxyDataRootPath, utils.WaitImportPrefix, fileName), responseBody, 0666)
	if err != nil {
		log.ErrorLog.Println(url, "将文件保存到本地错误:", err)
		return err
	}
	// 更新ProxySpiderTask表中的数据
	err = models.GormDB.Model(&models.ProxySpiderTask{}).Where("task_id = ?", taskId).Updates(map[string]interface{}{
		"status": 3,
	}).Error
	if err != nil {
		log.ErrorLog.Println(url, "更新ProxySpiderTask表中的数据错误:", err)
		return err
	}
	return nil
}

/*
serverAddr = "127.0.0.1"
serverPort = 7000

[[proxies]]
name = "test-tcp"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000

112.124.30.12:8211
*/
