package models

import (
	"database/sql"
	"time"
)

// BiliSpiderHistory b站抓取记录
type BiliSpiderHistory struct {
	Id             int64
	Key            string
	Value          string
	LastUpdateTime time.Time
}

func (b *BiliSpiderHistory) CreateTale() string {
	// 创建表
	return `CREATE TABLE IF NOT EXISTS bili_spider_history (
    				id INTEGER PRIMARY KEY AUTOINCREMENT,
    				key VARCHAR(255) unique,
    				value VARCHAR(255),
    				last_update_time DATETIME DEFAULT CURRENT_TIMESTAMP
                                               					)`
}

// GetDynamicBaseline 获取上次获取动态的最后baseline
func GetDynamicBaseline(db *sql.DB) string {
	r, err := db.Query("select value from bili_spider_history where `key` = 'dynamic_baseline'")
	if err != nil {
		return ""
	}
	var value string
	if r.Next() {
		r.Scan(&value)
	}
	r.Close()
	return value
}
func SaveDynamicBaseline(db *sql.DB, baseline string) {
	db.Exec("INSERT OR REPLACE INTO bili_spider_history (`key`,value) VALUES (?,?)", "dynamic_baseline", baseline)
}

func GetHistoryBaseLine(db *sql.DB) string {
	r, err := db.Query("select value from bili_spider_history where `key` = 'history_baseline'")
	if err != nil {
		return ""
	}
	var value string
	if r.Next() {
		r.Scan(&value)
	}
	r.Close()
	return value
}
func SaveHistoryBaseLine(db *sql.DB, baseline string) {
	db.Exec("INSERT OR REPLACE INTO bili_spider_history (`key`,value) VALUES (?,?)", "history_baseline", baseline)
}
