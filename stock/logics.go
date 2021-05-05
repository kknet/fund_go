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
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func GetSimpleStocks(codes []string) []map[string]interface{} { // 获取股票简略信息
	// 初始化
	results := make([]map[string]interface{}, len(codes))
	for i := range codes {
		// 读取redis
		info, err := rdb.HMGet(ctx, codes[i], "code", "name", "price", "pct_chg").Result()
		if err != nil {
			log.Println(err)
		}
		// 创建临时maps
		maps := map[string]interface{}{
			"code": info[0], "name": info[1], "price": info[2], "pct_chg": info[3],
		}
		// 尝试将maps中元素转成float
		for i := range maps {
			temp, err := strconv.ParseFloat(maps[i].(string), 2)
			if err != nil {
				continue
			}
			maps[i] = temp
		}
		results[i] = maps
	}
	return results
}

func GetDetailStocks(codes []string) []map[string]interface{} { // 获取股票详细信息
	// 初始化
	results := make([]map[string]interface{}, len(codes))
	for i := range codes {
		// 读取redis
		info, err := rdb.HGetAll(ctx, codes[i]).Result()
		if err != nil {
			log.Println(err)
		}
		// 创建临时maps
		maps := map[string]interface{}{}
		// 尝试将maps中元素转成float
		for i := range info {
			temp, err := strconv.ParseFloat(info[i], 2)
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
