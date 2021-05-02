package test

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

var ctx = context.Background()

// redis数据库
// 实时监听stock
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func GetDetailStock(code string) map[string]string { // 获取单只股票信息
	val, err := rdb.HGetAll(ctx, code).Result()
	//异常处理
	if err == redis.Nil {
		log.Println(code, "does not exist")
	} else if err != nil {
		panic(err)
	}
	return val
}

func GetSimpleStock(codes []string) []map[string]interface{} { // 获取多只简略信息
	result := make([]map[string]interface{}, 0)

	for i := range codes {
		code := codes[i]
		// 返回值为 []interface{}
		val, err := rdb.HMGet(ctx, code, "code", "name", "price", "pct_chg").Result()
		//异常处理
		if err == redis.Nil {
			continue
		} else if err != nil {
			panic(err)
		}
		result = append(result, map[string]interface{}{
			"code": val[0].(string), "name": val[1].(string), "price": val[2].(string), "pct_chg": val[3].(string),
		})
	}
	return result
}

func GetVolMaps(codes []string) map[string]string { //获取成交量字典
	result := map[string]string{}

	for i := range codes {
		code := codes[i]
		vol, err := rdb.HGet(ctx, code, "vol").Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			panic(err)
		}
		result[code] = vol
	}
	return result
}
