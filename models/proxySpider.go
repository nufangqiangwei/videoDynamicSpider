package models

import "time"

type ProxySpiderTask struct {
	Id             int64     `gorm:"primary_key" json:"id"`
	SpiderIp       string    `gorm:"varchar(255);notnull" json:"spider_ip"`
	TaskType       string    `gorm:"varchar(255);notnull" json:"task_type"`
	TaskName       string    `gorm:"varchar(255);notnull" json:"task_name"`
	ParamsSavePath string    `gorm:"varchar(255);notnull" json:"params_save_path"`
	TaskId         string    `gorm:"varchar(255);notnull" json:"task_id"`
	StartTimestamp time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"start_timestamp"`
	EndTimestamp   time.Time `gorm:"type:datetime" json:"end_timestamp"`
	Status         int       `gorm:"type:int;default:0" json:"status"` //-1:错误的任务 0:未开始 1:进行中 2:已完成
	ResultFileMd5  string    `gorm:"varchar(255)" json:"result_file_md5"`
}
