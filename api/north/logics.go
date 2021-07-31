package north

import (
	"github.com/go-gota/gota/dataframe"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"xorm.io/xorm"
)

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary
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
		return nil
	}
	return data
}

// GetPeriodData 获取阶段统计数据
func GetPeriodData(opt *PeriodOptions) []map[string]interface{} {
	// 时间
	var bulk *xorm.Session
	if opt.period > 0 {
		bulk = northDB.Table("agg_" + strconv.Itoa(opt.period) + "day")
	} else {
		bulk = northDB.Table(opt.tradeDate)
	}
	//sort
	orderName := opt.orderName
	if !opt.order {
		orderName = "-" + orderName
	}

	data, err := bulk.OrderBy(orderName).Limit(opt.size).QueryInterface()
	if err != nil {
		return nil
	}
	return data
}

// GetNorthFlow 北向资金流向
func GetNorthFlow() interface{} {
	url := "https://push2.eastmoney.com/api/qt/kamt.rtmin/get?fields1=f1,f3&fields2=f52,f54,f56"
	res, _ := http.Get(url)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	var str []string
	json.Get(body, "data", "s2n").ToVal(&str)

	df := dataframe.ReadCSV(strings.NewReader("hgt,sgt,all\n" + strings.Join(str, "\n")))
	return bson.M{
		"hgt": df.Col("hgt").Float(),
		"sgt": df.Col("sgt").Float(),
		"all": df.Col("all").Float(),
	}
}
