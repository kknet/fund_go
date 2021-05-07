package download

import (
	"fmt"
)

/* 计算行业涨跌幅 */
func setIndustry() {
	ids, _ := rdb.Keys(ctx, "industry:*").Result()
	// 根据行业名迭代
	for i := range ids {
		// 行业成分股
		//codes, _ := rdb.SMembers(ctx, ids[i]).Result()
		//codes
		fmt.Println(i)
	}
}

/* 计算地区涨跌幅 */
func setArea() {
	ids, _ := rdb.Keys(ctx, "area:*").Result()
	// 根据行业名迭代
	for i := range ids {
		// 行业成分股
		//codes, _ := rdb.SMembers(ctx, ids[i]).Result()
		//codes
		fmt.Println(i)
	}
}

/* 计算涨跌分布 */
func getNumber() {
	// 获取涨跌排行榜
	all := rdb.ZRangeWithScores(ctx, "rank:pct_chg", 0, -1).Val()
	numbers := []int64{
		// 跌停
		rdb.ZCount(ctx, "rank:委比", "-100", "-100").Val(),

		0, 0, 0, 0, 0, 0, 0, 0, 0,
		// 涨停
		rdb.ZCount(ctx, "rank:委比", "100", "100").Val(),
	}
	// 计算涨跌分布
	for _, mem := range all {
		pct := mem.Score

		if pct <= -7 {
			numbers[1]++
		} else if pct < -5 {
			numbers[2]++
		} else if pct <= -3 {
			numbers[3]++
		} else if pct < 0 {
			numbers[4]++
		} else if pct == 0 {
			numbers[5]++
		} else if pct < 3 {
			numbers[6]++
		} else if pct <= 5 {
			numbers[7]++
		} else if pct < 7 {
			numbers[8]++
		} else if pct >= 7 {
			numbers[9]++
		}
	}
	rdb.Del(ctx, "numbers")
	// 写入redis
	for _, num := range numbers {
		rdb.RPush(ctx, "numbers", num)
	}
}

func getMarketData() {
	go setIndustry()
	go setArea()
	go getNumber()
}
