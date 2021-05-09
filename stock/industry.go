package stock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
)

// redis数据库
var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0,
})

// GetIndustryData /* 获取行业数据 */
func GetIndustryData(idsType string) []map[string]interface{} {
	// idsType: industry, sw, area
	names, err := rdb.Keys(ctx, idsType+"*").Result()
	if err != nil {
		panic(err)
	}

	results := make([]map[string]interface{}, len(names))
	for i, name := range names {
		item, _ := rdb.HGetAll(ctx, name).Result()
		// 类型转换
		maps := map[string]interface{}{}
		for key, value := range item {
			temp, err := strconv.ParseFloat(value, 2)
			if err != nil {
				maps[key] = value
				continue
			}
			maps[key] = temp
		}
		results[i] = maps
	}
	return results
}

// GetNumbers /* 获取涨跌排行 */
func GetNumbers() map[string]interface{} {
	label, err := rdb.LRange(ctx, "numbers:label", 0, -1).Result()
	if err != nil {
		panic(err)
	}
	value, err := rdb.LRange(ctx, "numbers:value", 0, -1).Result()
	if err != nil {
		panic(err)
	}
	return map[string]interface{}{
		"label": label, "value": value,
	}
}
