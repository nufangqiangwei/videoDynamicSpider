package baseStruct

import (
	"os"
	"path/filepath"
)

const SqliteDaName = "videoInfo.db"

var RootPath = ""

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
