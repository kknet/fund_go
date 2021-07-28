package north

import (
	"log"
	"strconv"
	"xorm.io/xorm"
)

var northDB = ConnectDB()

// ConnectDB 连接数据库
func ConnectDB() *xorm.Engine {
	connStr := "postgres://postgres:123456@127.0.0.1:5432/north?sslmode=disable"
	db, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

// GetTopTen 获取十大成交股
func GetTopTen() []map[string]interface{} {
	data, err := northDB.Table("top_ten").QueryInterface()
	if err != nil {
		log.Println(err)
	}
	return data
}

// GetPeriodData 获取阶段统计数据
func GetPeriodData(opt *PeriodOptions) []map[string]interface{} {
	orderName := opt.orderName
	if !opt.order {
		orderName = "-" + orderName
	}

	data, err := northDB.
		Table("agg_" + strconv.Itoa(opt.period) + "day").
		OrderBy(orderName).
		Limit(opt.size).
		QueryInterface()

	if err != nil {
		return []map[string]interface{}{}
	}
	return data
}
