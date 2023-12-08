package utils

import (
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var (
	TimeWheelLog *log.Logger
	Info         *log.Logger
	Warning      *log.Logger
	ErrorLog     *log.Logger
	DBlog        *log.Logger
)

func InitLog(lofFilePath string) {
	var (
		logWriter      io.Writer
		dbWriter       io.Writer
		timeWheelWrite io.Writer
		dbLogFile      string
		timeWheeFile   string
	)
	if !strings.HasSuffix(lofFilePath, ".log") {
		dbLogFile = path.Join(lofFilePath, "db.log")
		lofFilePath = path.Join(lofFilePath, "videoSpider.log")
		timeWheeFile = path.Join(lofFilePath, "timeWheel.log")
	} else {
		rootPath, _ := path.Split(lofFilePath)
		dbLogFile = path.Join(rootPath, "db.log")
		lofFilePath = path.Join(rootPath, "videoSpider.log")
		timeWheeFile = path.Join(rootPath, "timeWheel.log")
	}

	println("日志文件路径：", lofFilePath)
	println("数据库日志文件路径：", dbLogFile)
	println("定时日志文件路径：", timeWheeFile)
	if runtime.GOOS == "linux" {
		logWriter, _ = rotateLogs.New(
			lofFilePath+".%Y-%m-%d",
			rotateLogs.WithLinkName(lofFilePath),
			rotateLogs.WithRotationTime(time.Hour*24),
		)
		dbWriter, _ = rotateLogs.New(
			dbLogFile+".%Y-%m-%d",
			rotateLogs.WithLinkName(dbLogFile),
			rotateLogs.WithRotationSize(50*1024*1024),
		)
		timeWheelWrite, _ = rotateLogs.New(
			timeWheeFile+".%Y-%m-%d",
			rotateLogs.WithLinkName(timeWheeFile),
			rotateLogs.WithRotationSize(50*1024*1024),
		)

	} else if runtime.GOOS == "windows" {
		logWriter, _ = os.OpenFile(lofFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		dbWriter, _ = os.OpenFile(dbLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		timeWheelWrite, _ = os.OpenFile(timeWheeFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)

	}

	log.SetOutput(logWriter)
	Info = log.New(logWriter, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(logWriter, "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(logWriter, "Error:", log.Ldate|log.Ltime|log.Lshortfile)

	DBlog = log.New(dbWriter, "DB:", log.Ldate|log.Ltime|log.Lshortfile)
	TimeWheelLog = log.New(timeWheelWrite, "定时:", log.Ldate|log.Ltime|log.Lshortfile)
}
