package baseStruct

import (
	"os"
	"path/filepath"
)

const SqliteDaName = "videoInfo.db"

var RootPath = "E:\\GoCode\\videoDynamicAcquisition\\cmd\\spider"

func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	if RootPath == "" {
		RootPath = filepath.Dir(ex)
	}

	println(RootPath)
}
