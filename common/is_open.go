package common

import "time"

// IsOpen 判断是否开市
func IsOpen(marketType string) bool {
	hour := time.Now().Hour()
	minute := time.Now().Minute()

	// CN
	if marketType == "CN" || marketType == "Index" {
		if time.Now().Weekday() <= 5 {
			// 上午
			if hour == 9 && minute >= 15 {
				return true
			} else if hour == 10 {
				return true
			} else if hour == 11 && minute < 30 {
				return true
			}
			//下午
			if 13 <= hour && hour < 15 {
				return true
			}
		}
		return false
		// HK
	} else if marketType == "HK" {
		if time.Now().Weekday() <= 5 {
			if 9 <= hour && hour < 12 {
				return true
			} else if 13 <= hour && hour < 16 {
				return true
			}
		}
		return false
		// US
	} else if marketType == "US" {
		if time.Now().Weekday() <= 5 && 21 <= hour {
			return true
		}
		if time.Now().Weekday() <= 5 && hour <= 6 {
			return true
		}
		return false
	}
	return false
}
