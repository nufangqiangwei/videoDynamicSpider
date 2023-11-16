package baseStruct

import (
	"os"
	"path/filepath"
)

const SqliteDaName = "videoInfo.db"

var RootPath = "C:\\Code\\GO\\videoDynamicSpider"

func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	if RootPath == "" {
		RootPath = filepath.Dir(ex)
	}

	println("RootPath地址：", RootPath)
}
