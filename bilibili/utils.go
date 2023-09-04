package bilibili

import (
	"strconv"
	"strings"
)

// HourAndMinutesAndSecondsToSeconds 时长字符串转秒数
// 时长字符串例子： [00:19, 02:45, 01:30:14]
func HourAndMinutesAndSecondsToSeconds(time string) int {
	split := strings.Split(time, ":")
	if len(split) == 2 {
		minutes, _ := strconv.Atoi(split[0])
		seconds, _ := strconv.Atoi(split[1])
		return minutes*60 + seconds
	} else if len(split) == 3 {
		hour, _ := strconv.Atoi(split[0])
		minutes, _ := strconv.Atoi(split[1])
		seconds, _ := strconv.Atoi(split[2])
		return hour*3600 + minutes*60 + seconds
	}
	return 0
}
