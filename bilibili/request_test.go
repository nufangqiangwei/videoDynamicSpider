package bilibili

import (
	"fmt"
	"testing"
	"time"
	"videoDynamicAcquisition/utils"
)

func TestHistory(t *testing.T) {
	utils.InitLog("E:\\GoCode\\videoDynamicAcquisition")
	fmt.Printf("%+v\n", getCollectVideoInfo(72121698, 1))
	time.Sleep(time.Second)
	fmt.Printf("%+v\n", getSeasonVideoInfo(1090255, 1))
}
