package baseStruct

import (
	"database/sql"
	"io"
	"os"
	"path"
	"sync"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

var dbFileLock sync.Mutex
var dbPath = path.Join(RootPath, SqliteDaName)

const backupDbPath = "/home/ubuntu/static/icon/videoInfo.db"

// CanUserDb 获取数据库连接。当前正在备份数据库时候不可用，直到备份结束
func CanUserDb() *sql.DB {
	if dbFileLock.TryLock() {
		dbFileLock.Unlock()
	} else {
		// 当前正在备份数据库,阻塞到备份结束
		dbFileLock.Lock()
		dbFileLock.Unlock()
	}
	db, _ := sql.Open("sqlite3", dbPath)
	return db
}

// BackUserDb 备份数据库
func BackUserDb() {
	dbFileLock.Lock()
	defer dbFileLock.Unlock()
	source, _ := os.Open(dbPath)
	defer source.Close()
	destination, _ := os.Create(backupDbPath)
	defer destination.Close()
	fineSize, err := io.Copy(destination, source)
	if err != nil {
		utils.ErrorLog.Println(err)
		return
	}
	utils.Info.Println("备份数据库成功，备份大小：", fineSize)
}

func InitDB() {
	utils.Info.Println(dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	baseModels := []models.BaseModel{&models.WebSite{}, &models.Author{}, &models.Video{}, &models.BiliAuthorVideoNumber{}, &models.BiliSpiderHistory{}, &models.VideoHistory{}}
	for _, baseModel := range baseModels {
		_, err = db.Exec(baseModel.CreateTale())
		if err != nil {
			utils.ErrorLog.Println("创建表失败")
			utils.ErrorLog.Println(err.Error())
			utils.ErrorLog.Println(baseModel.CreateTale())
		}
	}
	models.CreateDbLock(dbPath)

	db.Close()
}
