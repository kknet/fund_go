package download

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// redis数据库
var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0,
})

/* 更新全市场 排行榜 */
func ranks(stocks []map[string]interface{}) {
	for _, s := range stocks {
		// 去掉退市
		if s["总市值"].(float64) == 0 {
			continue
		}
		// 涨速
		member := redis.Z{Member: s["code"], Score: s["涨速"].(float64)}
		rdb.ZAdd(ctx, "rank:涨速", &member)
		// 5分钟涨幅
		member = redis.Z{Member: s["code"], Score: s["5min涨幅"].(float64)}
		rdb.ZAdd(ctx, "rank:5min涨幅", &member)
		// 涨幅
		member = redis.Z{Member: s["code"], Score: s["pct_chg"].(float64)}
		rdb.ZAdd(ctx, "rank:pct_chg", &member)
		// 每分钟更新
		if time.Now().Second() >= 54 {
			// 总市值
			member = redis.Z{Member: s["code"], Score: s["总市值"].(float64)}
			rdb.ZAdd(ctx, "rank:总市值", &member)
			// 主力净流入
			member = redis.Z{Member: s["code"], Score: s["主力净流入"].(float64)}
			rdb.ZAdd(ctx, "rank:主力净流入", &member)
			// 成交额
			member = redis.Z{Member: s["code"], Score: s["amount"].(float64)}
			rdb.ZAdd(ctx, "rank:amount", &member)
			// 委比
			member = redis.Z{Member: s["code"], Score: s["委比"].(float64)}
			rdb.ZAdd(ctx, "rank:委比", &member)
		}
	}
}
