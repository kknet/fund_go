package stock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
)

var ctx = context.Background()

// redis数据库
var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0,
})

func GetSimpleStocks(codes []string) []map[string]interface{} { // 获取股票简略信息
	// 初始化
	results := make([]map[string]interface{}, len(codes))

	for i, code := range codes {
		info, err := rdb.HMGet(ctx, "stock:"+code, "code", "name", "price", "pct_chg").Result()
		// 结果为空
		if info[0] == nil {
			info, err = rdb.HMGet(ctx, "index:"+code, "code", "name", "price", "pct_chg").Result()
		} else if err != nil {
			log.Println(err)
			continue
		} else if info[2] == nil {
			continue
		}
		// 类型转换
		price, _ := strconv.ParseFloat(info[2].(string), 2)
		pctChg, _ := strconv.ParseFloat(info[3].(string), 2)
		maps := map[string]interface{}{
			"code": info[0], "name": info[1], "price": price, "pct_chg": pctChg,
		}
		results[i] = maps
	}
	return results
}

func GetDetailStocks(codes []string) []map[string]interface{} { // 获取股票详细信息
	// 初始化
	results := make([]map[string]interface{}, len(codes))

	for i := range codes {
		info, err := rdb.HGetAll(ctx, "stock:"+codes[i]).Result()
		// 结果map为空
		if len(info) == 0 {
			info, err = rdb.HGetAll(ctx, "index:"+codes[i]).Result()
		} else if err != nil {
			log.Println(err)
		}
		// 创建临时maps
		maps := make(map[string]interface{}, len(info))
		// 尝试将maps中元素转成float
		for i := range info {
			temp, err := strconv.ParseFloat(info[i], 2)
			// 转换失败
			if err != nil {
				maps[i] = info[i]
				continue
			}
			maps[i] = temp
		}
		results[i] = maps
	}
	return results
}
