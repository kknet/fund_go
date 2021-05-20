package stock

import (
	"context"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"test/common"
)

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary
var ctx = context.Background()

// GetStockList 获取多只股票信息
func GetStockList(opt *common.CListOpt) []bson.M {
	var results []bson.M
	var err error
	// 指定Codes
	if opt.Codes[0] != "" {
		err = coll.Find(ctx, bson.M{"_id": bson.M{"$in": opt.Codes}}).Limit(30).All(&results)
		// Search
	} else if opt.Search != "" {
		match := bson.M{"$or": bson.A{
			// 正则匹配 不区分大小写
			bson.M{"_id": bson.M{"$regex": opt.Search, "$options": "i"}},
			bson.M{"name": bson.M{"$regex": opt.Search, "$options": "i"}},
		}}
		err = coll.Find(ctx, match).Limit(10).All(&results)
		fmt.Println(results)
		// RankList
	} else if opt.SortName != "" {

	}
	if err != nil {
		log.Println(err)
	}
	return results
}

// GetMinuteChart 获取分时行情
func GetMinuteChart(code string) bson.M {
	// 格式化代码为雪球格式
	code, err := FormatStock(code)
	if err != nil {
		return bson.M{"msg": err.Error()}
	}
	request := common.NewGetRequest("https://stock.xueqiu.com/v5/stock/chart/minute.json?period=1d&symbol="+code, true)
	client := http.Client{}
	res, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	str := json.Get(body, "data", "items").ToString()
	// json解析
	var temp []bson.M
	_ = json.Unmarshal([]byte(str), &temp)

	length := len(temp)
	price := make(bson.A, length)
	avg := make(bson.A, length)
	vol := make(bson.A, length)
	timestamp := make(bson.A, length)
	volCompare := make(bson.A, length)
	moneyFlow := make(bson.A, length)

	for i, x := range temp {
		price[i] = x["current"]
		vol[i] = x["volume"]
		avg[i] = x["avg_price"]
		timestamp[i] = x["timestamp"]
		volCompare[i] = bson.M{
			"now":  x["volume_compare"].(map[string]interface{})["volume_sum"],
			"last": x["volume_compare"].(map[string]interface{})["volume_sum_last"],
		}
		moneyFlow[i] = x["capital"]
	}
	return bson.M{
		"price": price, "vol": vol, "avg": avg, "timestamp": timestamp, "vol_compare": volCompare, "money_flow": moneyFlow,
	}
}

// Search 搜索股票
func Search(input string, searchType string) []bson.M {
	var results []bson.M
	match := bson.M{"$or": bson.A{
		// 正则匹配 不区分大小写
		bson.M{"_id": bson.M{"$regex": input, "$options": "i"}},
		bson.M{"name": bson.M{"$regex": input, "$options": "i"}},
	}}
	// 按成交量排序
	err := coll.Find(ctx, match).Limit(10).All(&results)
	if err != nil {
		log.Println(err)
	}
	return results
}

// GetNorthFlow 北向资金流向
func GetNorthFlow() {
	url := "https://push2.eastmoney.com/api/qt/kamt.rtmin/get?fields1=f1,f3&fields2=f52,f54,f56"
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(body)
}

// GetRank 全市场排行
func GetRank(marketType string) []bson.M {
	var results []bson.M
	err := coll.Find(ctx, bson.M{"marketType": marketType}).Limit(20).All(&results)
	if err != nil {
		log.Println(err)
	}
	return results
}

// GetRealtimeTicks 获取实时分笔成交
func GetRealtimeTicks(code string) []bson.M {
	// 格式化代码为雪球格式
	code, err := FormatStock(code)
	request := common.NewGetRequest("https://stock.xueqiu.com/v5/stock/history/trade.json?&count=30&symbol="+code, true)
	client := http.Client{}
	res, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	str := json.Get(body, "data", "items").ToString()
	// json解析
	var temp []bson.M
	_ = json.Unmarshal([]byte(str), &temp)

	results := make([]bson.M, len(temp))
	for i, item := range temp {
		results[i] = bson.M{
			"price":     item["current"],
			"type":      item["side"],
			"timestamp": item["timestamp"],
			"vol":       item["trade_volume"],
		}
	}
	return results
}

// FormatStock 股票代码格式化
// targetType: XeuQiu
func FormatStock(input string) (string, error) {
	item := strings.Split(input, ".")
	var code string

	if len(item) < 2 {
		return "", errors.New("代码格式不正确")
	}
	// 代码后缀
	switch item[1] {
	case "SH", "SZ":
		code = item[1] + item[0]
	case "HK", "US":
		code = item[0]
	default:
		return "", errors.New("代码格式不正确")
	}
	return code, nil
}
