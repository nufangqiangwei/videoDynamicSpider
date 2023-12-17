package models

import (
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
	"videoDynamicAcquisition/utils"
)

var (
	dbLock *utils.Flock
	GormDB *gorm.DB
)

func InitDB(dsn string) {
	cacheWebSite = make(map[string]WebSite)
	var err error
	GormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: logger.New(utils.DBlog, logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Info,
		}),
	})
	if err != nil {
		panic(err.Error())
	}
	//err = GormDB.AutoMigrate(&BiliSpiderHistory{}, &Author{}, &Video{}, &VideoAuthor{}, &VideoTag{}, &WebSite{},
	//	&Collect{}, &CollectVideo{}, &ProxySpiderTask{}, &Tag{}, &VideoHistory{}, &TaskToDoList{})
	if err != nil {
		println(err.Error())
		return
	}
	GormDB.Callback().Query().Before("gorm:query").Register("disable_raise_record_not_found", func(d *gorm.DB) {
		d.Statement.RaiseErrorOnNotFound = false
	})
	if err != nil {
		return
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
