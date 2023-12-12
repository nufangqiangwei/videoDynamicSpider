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
	SpiderIp       string     `gorm:"varchar(255);notnull" json:"spider_ip"`
	TaskType       string     `gorm:"varchar(255);notnull" json:"task_type"`
	TaskName       string     `gorm:"varchar(255);notnull" json:"task_name"`
	TaskId         string     `gorm:"varchar(255);notnull" json:"task_id"`
	StartTimestamp time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"start_timestamp"`
	EndTimestamp   *time.Time `gorm:"type:datetime" json:"end_timestamp"`
	Status         int        `gorm:"type:int;default:0" json:"status"` //-1:错误的任务 0:未开始 1:进行中 2:已完成 3:已下载
	ResultFileMd5  string     `gorm:"varchar(255)" json:"result_file_md5"`
}
