package real

import (
	"context"
	"errors"
	"fmt"
	"fund_go2/common"
	"fund_go2/download"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"math"
	"strconv"
	"strings"
)

const (
	pageSize = 20
)

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary
var ctx = context.Background()

// mongo collection
var coll = download.ConnectMongo("AllStock")

// GetStockList 获取多只股票信息
func GetStockList(codes []string) []map[string]interface{} {
	a := download.CNStock.Filter(
		dataframe.F{"codes", series.In, codes},
	)
	return a.Maps()
}

// AddSimpleMinute 添加简略分时行情
func AddSimpleMinute(items map[string]interface{}) {
	var info []string
	url := "https://push2.eastmoney.com/api/qt/stock/trends2/get?fields1=f1,f5,f8,f10,f11&fields2=f53&iscr=0&secid="
	url += items["cid"].(string) + "." + strings.Split(items["code"].(string), ".")[0]

	body, _ := common.NewGetRequest(url).Do()
	total := json.Get(body, "data", "trendsTotal").ToFloat32()
	timestamp := json.Get(body, "data", "time").ToInt()

	json.Get(body, "data", "trends").ToVal(&info)

	// 间隔
	space := 3
	results := make([]float64, 0)

	for i := 0; i < len(info); i += space {
		item := strings.Split(info[i], ",")[1]
		data, _ := strconv.ParseFloat(item, 8)
		results = append(results, data)
	}
	results = append(results, items["price"].(float64))

	items["chart"] = bson.M{
		"trends": results, "total": total, "timestamp": timestamp, "space": space,
	}
}

// Search 搜索股票
func search(input string) []bson.M {
	var results []bson.M
	match := bson.M{"$or": bson.A{
		// 正则匹配 不区分大小写
		bson.M{"_id": bson.M{"$regex": input, "$options": "i"}, "marketType": "CN", "type": "stock"},
		bson.M{"name": bson.M{"$regex": input, "$options": "i"}, "marketType": "CN", "type": "stock"},

		bson.M{"_id": bson.M{"$regex": input, "$options": "i"}, "marketType": "HK"},
		bson.M{"name": bson.M{"$regex": input, "$options": "i"}, "marketType": "HK"},

		bson.M{"_id": bson.M{"$regex": input, "$options": "i"}, "marketType": "US"},
		bson.M{"name": bson.M{"$regex": input, "$options": "i"}, "marketType": "US"},
	}}
	_ = coll.Find(ctx, match).Limit(12).All(&results)
	return results
}

// getRank 全市场排行
func getRank(opt *common.RankOpt) []map[string]interface{} {
	indexes := make([]int, pageSize)

	for i := 0; i < pageSize; i++ {
		indexes[i] = (opt.Page-1)*pageSize + i
	}

	data := download.CNStock.Arrange(
		dataframe.RevSort(opt.SortName),
	).Subset(indexes)

	return data.Maps()
}

// PanKou  获取五档明细
func PanKou(code string) bson.M {
	// 格式化代码为雪球格式
	code, err := formatStock(code)
	if err != nil {
		return bson.M{"msg": "代码格式错误"}
	}
	url := "https://stock.xueqiu.com/v5/stock/realtime/pankou.json?&symbol=" + code
	body, err := common.NewGetRequest(url, true).Do()
	if err != nil {
		panic(err)
	}
	str := json.Get(body, "data").ToString()
	// json解析
	var data bson.M
	_ = json.Unmarshal([]byte(str), &data)
	return data
}

// GetRealtimeTicks 获取实时分笔成交
func GetRealtimeTicks(code string) bson.M {
	// 格式化代码为雪球格式
	code, err := formatStock(code)
	if err != nil {
		return bson.M{"msg": "代码格式错误"}
	}
	url := "https://stock.xueqiu.com/v5/stock/history/trade.json?&count=50&symbol=" + code
	body, err := common.NewGetRequest(url, true).Do()
	if err != nil {
		return bson.M{"msg": err.Error()}
	}
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
	return bson.M{"data": results}
}

// formatStock 股票代码格式化为雪球代码
func formatStock(input string) (string, error) {
	item := strings.Split(input, ".")
	var code string

	if len(item) < 2 {
		return "", errors.New("代码格式不正确")
	}
	// 代码后缀
	switch item[1] {
	case "SH", "SZ", "sh", "sz":
		code = item[1] + item[0]
	case "HK", "US", "us", "hk":
		code = item[0]
	default:
		return "", errors.New("代码格式不正确")
	}
	return code, nil
}

// getNumbers 获取涨跌分布
func getNumbers(marketType string) bson.M {

	var temp []bson.M
	_ = coll.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"marketType": marketType, "type": "stock"}}},
		bson.D{{"$group", bson.M{
			"_id":     nil,
			"pct_chg": bson.M{"$push": "$pct_chg"},
			"wb":      bson.M{"$push": "$wb"},
		}}},
	}).All(&temp)

	res := make([]int32, 11)
	pct := temp[0]["pct_chg"].(bson.A)
	wb := temp[0]["wb"].(bson.A)

	for i := range temp {
		p := pct[i].(float64) //涨跌幅pct_chg
		w := wb[i].(float64)

		if p < -7 {
			res[1]++
		} else if p < -5 {
			res[2]++
		} else if p < -3 {
			res[3]++
		} else if p < 0 {
			res[4]++
		} else if p == 0 {
			res[5]++
		} else if p <= 3 {
			res[6]++
		} else if p <= 5 {
			res[7]++
		} else if p <= 7 {
			res[8]++
		} else if p > 7 {
			res[9]++
		}

		if marketType != "CN" {
			if p < -10 {
				res[0]++
			} else if p > 10 {
				res[10]++
			}
		} else {
			if w == -100 {
				res[0]++
			} else if w == 100 {
				res[10]++
			}
		}
	}
	label := []string{"跌停", "<7", "7-5", "5-3", "3-0", "0", "0-3", "3-5", "5-7", ">7", "涨停"}
	if marketType != "CN" {
		label[0] = "<10"
		label[10] = ">10"
	}
	return bson.M{"label": label, "value": res}
}

// getIndustry 获取板块行情
// marketType=CN; name=sw, industry, area
func getIndustry(name string) []bson.M {
	var results []bson.M
	_ = coll.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"marketType": "CN", "type": "stock", name: bson.M{"$nin": bson.A{math.NaN(), nil, ""}}}}},
		bson.D{{"$sort", bson.M{"pct_chg": -1}}},
		bson.D{{"$group", bson.M{
			"_id":         "$" + name,
			"max_pct":     bson.M{"$first": "$pct_chg"},
			"count":       bson.M{"$sum": 1},
			"领涨股":         bson.M{"$first": "$name"},
			"主力净流入":       bson.M{"$sum": "$main_net"},
			"mc":          bson.M{"$sum": "$mc"},
			"pe_ttm":      bson.M{"$avg": "$pe_ttm"},
			"pb":          bson.M{"$avg": "$pb"},
			"vol":         bson.M{"$sum": "$vol"},
			"amount":      bson.M{"$sum": "$amount"},
			"total_share": bson.M{"$sum": "$total_share"},
			"power":       bson.M{"$sum": bson.M{"$multiply": bson.A{"$mc", "$pct_chg"}}},
		}}},
	}).All(&results)
	// 由权重计算涨跌幅
	for _, i := range results {
		i["pct_chg"] = i["power"].(float64) / i["mc"].(float64)
		// 换手率
		i["tr"] = i["vol"].(float64) / i["total_share"].(float64) * 10000
		delete(i, "power")
		delete(i, "total_share")
	}
	return results
}

// getNorthFlow 北向资金流向
func getNorthFlow() {
	url := "https://push2.eastmoney.com/api/qt/kamt.rtmin/get?fields1=f1,f3&fields2=f52,f54,f56"
	body, err := common.NewGetRequest(url).Do()
	if err != nil {
		log.Println(err)
	}
	str := json.Get(body, "data", "s2n").ToString()
	// json解析
	var temp []string
	_ = json.Unmarshal([]byte(str), &temp)
	fmt.Println(temp)
}