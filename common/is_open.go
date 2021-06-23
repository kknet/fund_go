package common

import (
	_ "github.com/lib/pq"
	"time"
)

// IsOpen 判断是否开市
func IsOpen(marketType string) bool {
	// CN
	if marketType == "CN" || marketType == "CNIndex" {
		if marketType == "CN" {
			return true
		}
		if time.Now().Weekday() <= 5 {
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
		// HK
	} else if marketType == "HK" {
		if time.Now().Weekday() <= 5 {
			if 9 <= time.Now().Hour() && time.Now().Hour() < 12 {
				return true
			} else if 13 <= time.Now().Hour() && time.Now().Hour() < 16 {
				return true
			}
		}
		return false
		// US
	} else if marketType == "US" {
		if time.Now().Weekday() <= 5 && 21 <= time.Now().Hour() {
			return true
		}
		if time.Now().Weekday() == 5 && time.Now().Hour() <= 6 {
			return true
		}
		return false
	}
	return false
}
