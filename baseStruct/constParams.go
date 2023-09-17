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
	RootPath = filepath.Dir(ex)
}
