package stock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

var ctx = context.Background()

// redis数据库
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func GetSimpleStocks(codes []string) []map[string]string { // 获取股票简略信息

	results := make([]map[string]string, 0)
	for i := range codes {
		info, err := rdb.HMGet(ctx, codes[i], "code", "name", "price", "pct_chg").Result()
		if err != nil {
			log.Println(err)
		}
		results = append(results, map[string]string{
			"code": info[0].(string), "name": info[1].(string), "price": info[2].(string), "pct_chg": info[3].(string),
		})
	}
	return results
}

func GetDetailStocks(codes []string) []map[string]string { // 获取股票详细信息

	results := make([]map[string]string, 0)
	for i := range codes {
		info, err := rdb.HGetAll(ctx, codes[i]).Result()
		if err != nil {
			log.Println(err)
		}
		results = append(results, info)
	}
	return results
}
