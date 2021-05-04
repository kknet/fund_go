package marketime

import "time"

func IsOpen() bool { //判断是否开市

	if time.Now().Weekday() < 5 {
		// 上午
		if time.Now().Hour() == 9 && time.Now().Minute() >= 15 {
			return true
		} else if time.Now().Hour() == 10 {
			return true
		} else if time.Now().Hour() == 11 && time.Now().Minute() < 30 {
			return true
		}
		//下午
		if 13 <= time.Now().Hour() && time.Now().Hour() < 15 {
			return true
		}
	}
	return false
}
