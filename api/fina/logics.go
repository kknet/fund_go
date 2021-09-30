package fina

import (
	"fund_go2/env"
	"xorm.io/xorm"
)

var finaDB *xorm.Engine

func init() {
	var err error

	connStr := "postgres://postgres:123456@" + env.PostgresHost + "/fund?sslmode=disable"
	finaDB, err = xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
}

// GetFinaData 获取股票财务数据
func GetFinaData(code string, period string) interface{} {

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

	data, _ := finaDB.Table("fina").
		Where("ts_code=?", code).
		Where("trade_date like ?", "%"+endDate).
		OrderBy("trade_date").
		QueryString()
	return data
}

// FilterStock 指标选股
func FilterStock() interface{} {
	info, _ := finaDB.Table("agg").
		Where("roe >= 20").Where("roa >= 10").Where("grossprofit_margin >= 25").
		Where("netprofit_yoy >= 20").Where("op_yoy >= 10").Where("or_yoy >= 10").
		Where("pe_ttm <= 50").Where("total_mv >= 1000000").Where("now_n_income = max_n_income").QueryString()
	return info
}
