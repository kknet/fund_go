package fina

import (
	"go.mongodb.org/mongo-driver/bson"
	"xorm.io/xorm"
)

var FinaDB = ConnectDB()

// ConnectDB 连接数据库
func ConnectDB() *xorm.Engine {
	connStr := "postgres://postgres:123456@127.0.0.1:5432/fina?sslmode=disable"
	db, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

// GetFinaData 获取股票财务数据
func GetFinaData(code string) interface{} {
	return code
}

// FilterStock 指标选股
func FilterStock() interface{} {
	data := bson.M{}
	info, _ := FinaDB.Table("agg").
		Where("roe >= 20").Where("roa >= 10").Where("grossprofit_margin >= 25").
		Where("netprofit_yoy >= 20").Where("op_yoy >= 10").Where("or_yoy >= 10").
		Where("pe_ttm <= 50").Where("total_mv >= 1000000").Where("now_n_income = max_n_income").QueryInterface()

	data["count"] = len(info)
	data["data"] = info
	return data
}
