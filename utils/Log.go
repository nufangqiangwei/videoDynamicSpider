package utils

import (
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"strings"
	"time"
)

var (
	writer       *rotateLogs.RotateLogs
	TimeWheelLog *log.Logger
	Info         *log.Logger
	Warning      *log.Logger
	ErrorLog     *log.Logger
)

func InitLog(lofFilePath string) {
	if strings.HasSuffix(lofFilePath, ".log") {
		lofFilePath = lofFilePath
	} else {
		lofFilePath = lofFilePath + "SSO.log"
	}

	writer, _ = rotateLogs.New(
		lofFilePath+".%Y-%m-%d",
		rotateLogs.WithLinkName(lofFilePath),
		rotateLogs.WithMaxAge(time.Hour*24*30),
		rotateLogs.WithRotationTime(time.Hour*24),
	)
	log.SetOutput(writer)
	TimeWheelLog = log.New(writer, "定时:", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(writer, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(writer, "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(writer, "Error:", log.Ldate|log.Ltime|log.Lshortfile)
}
