package download

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var db = openDB()

// 连接数据库
func openDB() *sql.DB {
	db, err := sql.Open("sqlite3", "C:/Users/lucario/PycharmProjects/fund/stock.db")
	if err != nil {
		panic(err)
	}
	return db
}

// GetIndustry 获取行业 marketType=CN, US, HK
func GetIndustry(marketType string) {
	var sqlStr string

	if marketType == "CN" {
		sqlStr = `select
		sw as name,
		name as 领涨股,
		max(pct_chg) as max_pct,
		sum(total_share * price) / 10000 as 总市值,
		sum(float_share * price) / 10000 as 流通市值,
		avg(pe_ttm) as pe_ttm,
		avg(pb) as pb,
		sum(vol) as vol,
		sum(amount) as amount,
		sum(vol) / sum(total_share) as 换手率,
		sum(主力净流入) / 100000000 as 主力净流入,
		sum(主力净流入) / (sum(主力流入) + sum(主力流出)) * 100 as jlr_pct` +
			"\nfrom " + marketType + "Basic," + marketType + "Stock" +
			"\nwhere " + marketType + "Basic.code=" + marketType + "Stock.code and sw is not null" +
			"\ngroup by sw order by pct_chg desc"
	} else {
		sqlStr = `select
		sw as name,
		name as 领涨股,
		max(pct_chg) as max_pct,
		sum(total_share * price) / 10000 as 总市值,
		sum(float_share * price) / 10000 as 流通市值,
		avg(pe_ttm) as pe_ttm,
		avg(pb) as pb,
		sum(vol) as vol,
		sum(amount) as amount,
		sum(vol) / sum(total_share) as 换手率,
		sum(主力净流入) / 100000000 as 主力净流入,
		sum(主力净流入) / (sum(主力流入) + sum(主力流出)) * 100 as jlr_pct` +
			"\nfrom " + marketType + "Basic," + marketType + "Stock" +
			"\nwhere " + marketType + "Basic.code=" + marketType + "Stock.code and sw is not null" +
			"\ngroup by sw order by pct_chg desc"
	}
	rows, _ := db.Query(sqlStr)
	for rows.Next() {
	}
}

// GetNumbers 获取市场涨跌分布
// marketType = CN, US, HK
func GetNumbers(marketType string) map[string]interface{} {
	var sqlStr string
	// 中国股市
	if marketType == "CN" {
		sqlStr = `select
			sum(委比 == -100) as 跌停,
			sum(pct_chg < -7) as '<7',
			sum(-7 <= pct_chg and pct_chg < -5) as '7-5',
			sum(-5 <= pct_chg and pct_chg < -3) as '5-3',
			sum(-3 <= pct_chg and pct_chg < 0) as '3-0',
			sum(pct_chg == 0) as '0',
			sum(0 < pct_chg and pct_chg <= 3) as '0-3',
			sum(3 < pct_chg and pct_chg <= 5) as '3-5',
			sum(5 < pct_chg and pct_chg <= 7) as '5-7',
			sum(pct_chg > 7) as '>7',
			sum(委比 == 100) as 涨停` +
			"\nfrom " + marketType + "Basic," + marketType + "Stock" +
			"\nwhere " + marketType + "Basic.code=" + marketType + "Stock.code"
		// 外盘
	} else if marketType == "US" || marketType == "HK" {
		sqlStr = `select
			sum(pct_chg < -10) as '<10',
			sum(pct_chg < -7) as '<7',
			sum(-7 <= pct_chg and pct_chg < -5) as '7-5',
			sum(-5 <= pct_chg and pct_chg < -3) as '5-3',
			sum(-3 <= pct_chg and pct_chg < 0) as '3-0',
			sum(pct_chg == 0) as '0',
			sum(0 < pct_chg and pct_chg <= 3) as '0-3',
			sum(3 < pct_chg and pct_chg <= 5) as '3-5',
			sum(5 < pct_chg and pct_chg <= 7) as '5-7',
			sum(pct_chg > 10) as '>10'` +
			"\nfrom " + marketType + "Stock"
		// 不存在
	} else {
		return map[string]interface{}{}
	}
	rows, _ := db.Query(sqlStr)
	rows.Next()
	var r0, r1, r2, r3, r4, r5, r6, r7, r8, r9, r10 int

	_ = rows.Scan(&r0, &r1, &r2, &r3, &r4, &r5, &r6, &r7, &r8, &r9, &r10)
	fmt.Println(r0, r1, r2, r3, r4, r5)

	s, _ := rows.Columns()
	maps := map[string]interface{}{
		"label": s,
		"value": []int{r0, r1, r2, r3, r4, r5, r6, r7, r8, r9, r10},
	}
	return maps
}
