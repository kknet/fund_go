package download

import (
	"github.com/go-redis/redis/v8"
	"time"
)

/* 更新全市场 排行榜 */
func ranks(stocks []map[string]interface{}) {
	for i := range stocks {
		s := stocks[i]
		// 去掉退市
		if s["总市值"].(float64) == 0 {
			continue
		}
		// 涨速
		member := redis.Z{Member: s["code"], Score: s["涨速"].(float64)}
		rdb.ZAdd(ctx, "rank:up_speed", &member)
		// 5分钟涨幅
		member = redis.Z{Member: s["code"], Score: s["5min涨幅"].(float64)}
		rdb.ZAdd(ctx, "rank:5min_pct", &member)
		// 涨幅
		member = redis.Z{Member: s["code"], Score: s["pct_chg"].(float64)}
		rdb.ZAdd(ctx, "rank:pct_chg", &member)
		// 每分钟更新
		if time.Now().Second() != 999 {
			// 市值
			member = redis.Z{Member: s["code"], Score: s["总市值"].(float64)}
			rdb.ZAdd(ctx, "rank:market_value", &member)
			// 主力净流入
			member = redis.Z{Member: s["code"], Score: s["主力净流入"].(float64)}
			rdb.ZAdd(ctx, "rank:main_net", &member)
			// 成交额
			member = redis.Z{Member: s["code"], Score: s["amount"].(float64)}
			rdb.ZAdd(ctx, "rank:amount", &member)
		}
	}
}
