package test

import (
	"context"
	"fmt"
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

func GetDetailStock(code string) map[string]string { // 获取股票信息
	val, err := rdb.HGetAll(ctx, code).Result()
	//异常处理
	if err == redis.Nil {
		log.Println(code, "does not exist")
	} else if err != nil {
		panic(err)
	}
	return val
}

func GetSimpleStock(code string) map[string]interface{} { // 获取简略信息
	// 返回值为 []interface{}
	val, err := rdb.HMGet(ctx, code, "code", "name", "price", "pct_chg").Result()
	//异常处理
	if err == redis.Nil {
		log.Println(code, "does not exist")
	} else if err != nil {
		panic(err)
	}
	temp := map[string]interface{}{"code": val[0], "name": val[1], "price": val[2], "pct_chg": val[3]}
	fmt.Println(temp)
	fmt.Println("返回结果：", temp)
	return temp
}

// 输入输出都为[]map，key为代码，value为成交量，当value不同时返回该key &value

func CheckUpdate(data map[string]float32) map[string]float32 {
	result := make(map[string]float32, len(data))
	// 迭代map中的key
	for code := range data {
		vol, err := rdb.HMGet(ctx, code, "vol").Result()
		if err == redis.Nil {
			log.Println(code, "does not exist")
		} else if err != nil {
			panic(err)
		}
		result = map[string]float32{"code": vol[0].(float32)}
	}
	return result
}

func GetVol(code string) (string, error) { //获取股票成交量
	vol, err := rdb.HGet(ctx, code, "vol").Result()
	if err == redis.Nil {
		return "", err
	} else if err != nil {
		return "", err
	}
	return vol, nil
}
