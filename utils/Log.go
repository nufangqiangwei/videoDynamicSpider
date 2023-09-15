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
)

func InitLog(lofFilePath string) {
	var writer io.Writer
	if !strings.HasSuffix(lofFilePath, ".log") {
		lofFilePath = path.Join(lofFilePath, "videoSpider.log")
	}
	if runtime.GOOS == "linux" {
		writer, _ = rotateLogs.New(
			lofFilePath+".%Y-%m-%d",
			rotateLogs.WithLinkName(lofFilePath),
			rotateLogs.WithMaxAge(time.Hour*24*30),
			rotateLogs.WithRotationTime(time.Hour*24),
		)
	} else if runtime.GOOS == "windows" {
		writer, _ = os.OpenFile(lofFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	}

	log.SetOutput(writer)
	TimeWheelLog = log.New(writer, "定时:", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(writer, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(writer, "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(writer, "Error:", log.Ldate|log.Ltime|log.Lshortfile)
}
