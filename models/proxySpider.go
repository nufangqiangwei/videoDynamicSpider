package models

import (
	"gorm.io/gorm"
	"time"
)

// TaskToDoList 待爬取的任务表
type TaskToDoList struct {
	gorm.Model
	TaskType   string `gorm:"varchar(255);notnull" json:"task_type"`
	TaskParams string `gorm:"varchar(255);notnull" json:"task_params"`
	Status     int    `gorm:"type:int;default:0" json:"status"`      //-1:错误的任务 0:未开始 1:进行中 2:已完成
	RunTaskId  uint   `gorm:"type:int;default:0" json:"run_task_id"` //对应的爬虫任务id,ProxySpiderTask表的id
}

// ProxySpiderTask 代理爬虫任务表
type ProxySpiderTask struct {
	gorm.Model
	SpiderIp       string    `gorm:"varchar(255);notnull" json:"spider_ip"`
	TaskType       string    `gorm:"varchar(255);notnull" json:"task_type"`
	TaskName       string    `gorm:"varchar(255);notnull" json:"task_name"`
	TaskId         string    `gorm:"varchar(255);notnull" json:"task_id"`
	StartTimestamp time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"start_timestamp"`
	EndTimestamp   time.Time `gorm:"type:datetime" json:"end_timestamp"`
	Status         int       `gorm:"type:int;default:0" json:"status"` //-1:错误的任务 0:未开始 1:进行中 2:已完成
	ResultFileMd5  string    `gorm:"varchar(255)" json:"result_file_md5"`
}

/*
使用golang代码写一个定时执行的函数，该功能使用的两张表是这两个，数据库操作使用gorm，目前已初始化好DB对象，参数名是GormDB。
函数功能描述如下：
定时查询TaskToDoList表，按照TaskType进行聚合，每100条任务进行一次分配，分配到代理上进行执行。
分配的时候，如果这个代理目前正在有执行的任务就换下一个，如果当前都在进行任务，就放弃分配，等待下次循环在进行分配。
任务被分配后，这100行数据的Status需要更新未1，表示正在进行中。
任务分配通过http POST请求.请求地址为http://{SpiderIp}/{TaskType} 请求类型为json类型，body格式{"IdList":[]string{}},列表中未100个任务的TaskParams
接口会返回一个uuid储存到 ProxySpiderTask.TaskId字段中。
*/
