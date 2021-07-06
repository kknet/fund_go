package fina

import (
	"go.mongodb.org/mongo-driver/bson"
	"xorm.io/xorm"
)

var finaDB = ConnectDB()

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
func GetFinaData(code string, period string) bson.M {
	data := bson.M{}

	var endDate string
	switch period {
	case "q", "1q":
		endDate = "0331"
	case "2q":
		endDate = "0630"
	case "3q":
		endDate = "0930"
	case "4q", "y":
		endDate = "1231"
	}
	// 添加每时期数据
	for _, year := range []string{"2021", "2020", "2019", "2018", "2017", "2016"} {
		data[year+endDate], _ = finaDB.Table(year+endDate).Where("ts_code = ?", code).QueryInterface()
	}
	// 添加agg复合数据
	data["agg"], _ = finaDB.Table("agg").Where("ts_code = ?", code).QueryInterface()
	return data
}

// FilterStock 指标选股
func FilterStock() interface{} {
	data := bson.M{}
	info, _ := finaDB.Table("agg").
		Where("roe >= 20").Where("roa >= 10").Where("grossprofit_margin >= 25").
		Where("netprofit_yoy >= 20").Where("op_yoy >= 10").Where("or_yoy >= 10").
		Where("pe_ttm <= 50").Where("total_mv >= 1000000").Where("now_n_income = max_n_income").QueryInterface()

	data["count"] = len(info)
	data["data"] = info
	return data
}
