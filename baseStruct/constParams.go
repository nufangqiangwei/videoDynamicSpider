package baseStruct

import (
	"os"
	"path/filepath"
)

var RootPath = "E:\\GoCode\\videoDynamicAcquisition"

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
