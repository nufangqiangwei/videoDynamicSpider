package utils

import (
	"fmt"
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
	Info     *log.Logger
	Warning  *log.Logger
	ErrorLog *log.Logger
)

type LogInputFile struct {
	FileName     string
	File         io.Writer
	WriterObject *log.Logger
}

func InitLog(logPath string, outPutFile ...string) []LogInputFile {
	var (
		logWriter   io.Writer
		lofFilePath string
		rootPath    string
	)
	if !strings.HasSuffix(logPath, ".log") {
		lofFilePath = path.Join(logPath, "videoSpider.log")
		rootPath = logPath
	} else {
		rootPath, _ = path.Split(logPath)
		lofFilePath = path.Join(rootPath, "videoSpider.log")
	}

	println("日志文件路径：", lofFilePath)

	logWriter = createLogFile(lofFilePath)
	log.SetOutput(logWriter)
	Info = log.New(logWriter, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(logWriter, "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(logWriter, "Error:", log.Ldate|log.Ltime|log.Lshortfile)

	result := make([]LogInputFile, 0)
	for _, logFileName := range outPutFile {
		f := path.Join(rootPath, fmt.Sprintf("%s.log", logFileName))
		println(logFileName, "日志文件路径：", f)
		logFile := LogInputFile{
			FileName: logFileName,
			File:     createLogFile(f),
		}
		logFile.WriterObject = log.New(logFile.File, "", log.Ldate|log.Ltime|log.Lshortfile)
		result = append(result, logFile)
	}
	return result
}

func createLogFile(logfilePath string) io.Writer {
	var logWriter io.Writer
	if runtime.GOOS == "linux" {
		logWriter, _ = rotateLogs.New(
			logfilePath+".%Y-%m-%d",
			rotateLogs.WithLinkName(logfilePath),
			rotateLogs.WithRotationTime(time.Hour*24),
		)
	} else if runtime.GOOS == "windows" {
		logWriter, _ = os.OpenFile(logfilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	}
	return logWriter
}
