package main

import (
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"strings"
	"time"
)

var (
	writer       *rotateLogs.RotateLogs
	timewheelLog *log.Logger
	info         *log.Logger
	warning      *log.Logger
	errorLog     *log.Logger
	lofFilePath  string
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
	timewheelLog = log.New(writer, "定时:", log.Ldate|log.Ltime|log.Lshortfile)
	info = log.New(writer, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	warning = log.New(writer, "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(writer, "Error:", log.Ldate|log.Ltime|log.Lshortfile)
}
