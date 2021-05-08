package download

import (
	"fmt"
)

/* 从redis中初始化 */
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
