package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
	"videoDynamicAcquisition/utils"
)

var (
	dbLock *utils.Flock
	GormDB *gorm.DB
)

func InitDB(dsn string, createModel bool, log *log.Logger) {
	cacheWebSite = make(map[string]WebSite)
	var (
		err    error
		Logger logger.Interface
	)
	if log != nil {
		Logger = logger.New(log, logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Warn,
		})
	}

	GormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: Logger,
	})
	if err != nil {
		panic(err.Error())
	}
	if createModel {
		err = GormDB.AutoMigrate(&BiliSpiderHistory{}, &Author{}, &Video{}, &VideoAuthor{}, &VideoTag{}, &WebSite{},
			&Collect{}, &CollectVideo{}, &ProxySpiderTask{}, &Tag{}, &VideoHistory{}, &TaskToDoList{}, &VideoPlayData{})
		if err != nil {
			println(err.Error())
			return
		}
	}

	GormDB.Callback().Query().Before("gorm:query").Register("disable_raise_record_not_found", func(d *gorm.DB) {
		d.Statement.RaiseErrorOnNotFound = false
	})
	if err != nil {
		panic(err.Error())
	}
}

func timeCheck(t time.Time) interface{} {
	var followTime interface{} = nil
	if t.Unix() > 0 {
		followTime = t.Format("2006-01-02 15:04:05")
	}
	return followTime
}

type BaseModel interface {
	CreateTale() string
}

func CreateDbLock(dbPath string) {
	if dbLock == nil {
		dbLock = utils.NewFlock(dbPath)
	} else {
		println("dbLock已经存在")
	}
}

type CustomTime struct {
	time.Time
}

func (t CustomTime) MarshalJSON() ([]byte, error) {
	formatted := t.Format("2006-01-02 15:04:05")
	return []byte(`"` + formatted + `"`), nil
}

func (t *CustomTime) UnmarshalJSON(data []byte) error {
	parsed, err := time.Parse(`"2006-01-02 15:04:05"`, string(data))
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}
