package bilibili

import (
	"strconv"
	"testing"
)

func TestModifyVideoPushTime(t *testing.T) {
	var (
		Baseline string
		IdStr    interface{}
	)
	IdStr = 1243567865432345
	Baseline, ok := IdStr.(string)
	if !ok {
		a, ok := IdStr.(int)
		if ok {
			Baseline = strconv.Itoa(a)
		}
	}
	println(Baseline)
}
