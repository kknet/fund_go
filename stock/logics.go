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
	"strconv"
	"strings"
	"test/common"
)

type minute struct {
	Time    string  `json:"time"`
	Price   float64 `json:"price"`
	Vol     float64 `json:"vol"`
	Amount  float64 `json:"amount"`
	Net     float64 `json:"net"`      // 资金净流入
	MainNet float64 `json:"main_net"` // 大单净流入
}

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary
var ctx = context.Background()

// GetStockList 获取多只股票信息
func GetStockList(codes []string) []bson.M {
	var results []bson.M
	err := coll.Find(ctx, bson.M{"_id": bson.M{"$in": codes}}).Limit(50).All(&results)
	if err != nil {
		log.Println(err)
	}
	return results
}

// AddSimpleMinute 添加简略分时行情
func AddSimpleMinute(items bson.M) {
	// 获取cid
	cid := int(items["cid"].(float64))
	symbol := strings.Split(items["code"].(string), ".")[0]

	var info []string
	url := "https://push2.eastmoney.com/api/qt/stock/trends2/get?fields1=f1,f5,f8,f10,f11&fields2=f53&iscr=0&secid="
	url += strconv.Itoa(cid) + "." + symbol

	body, _ := common.NewGetRequest(url).Do()
	preClose := json.Get(body, "data", "preClose").ToFloat32()
	total := json.Get(body, "data", "trendsTotal").ToFloat32()
	timestamp := json.Get(body, "data", "time").ToInt()

	json.Get(body, "data", "trends").ToVal(&info)

	// 间隔
	space := 2
	results := make([]float64, 0)

	for i := len(info) % space; i < len(info); i += space {
		item := strings.Split(info[i], ",")[1]
		data, _ := strconv.ParseFloat(item, 8)
		results = append(results, data)
	}
	items["chart"] = bson.M{
		"trends": results, "close": preClose, "total": total, "timestamp": timestamp, "space": space,
	}
}

// Search 搜索股票
func Search(input string, marketType string) interface{} {
	var results []bson.M
	// 先搜索CN
	match := bson.M{"$or": bson.A{
		// 正则匹配 不区分大小写
		bson.M{"_id": bson.M{"$regex": input, "$options": "i"}, "marketType": marketType, "type": "stock"},
		bson.M{"name": bson.M{"$regex": input, "$options": "i"}, "marketType": marketType, "type": "stock"},
	}}
	// 按成交量排序
	err := coll.Find(ctx, match).Limit(12).All(&results)
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
func GetRank(opt *common.RankOpt) []bson.M {
	var results []bson.M
	err := coll.Find(ctx, bson.M{"marketType": opt.MarketType, "type": "stock"}).Limit(20).All(&results)
	if err != nil {
		log.Println(err)
	}
	return results
}

// GetPanKou  获取五档明细
func GetPanKou(code string) bson.M {
	// 格式化代码为雪球格式
	code, err := FormatStock(code)
	url := "https://stock.xueqiu.com/v5/stock/realtime/pankou.json?&symbol=" + code
	body, err := common.NewGetRequest(url, true).Do()
	if err != nil {
		panic(err)
	}
	str := json.Get(body, "data").ToString()
	// json解析
	var temp bson.M
	_ = json.Unmarshal([]byte(str), &temp)
	return temp
}

// GetRealtimeTicks 获取实时分笔成交
func GetRealtimeTicks(code string) bson.M {
	// 格式化代码为雪球格式
	code, err := FormatStock(code)

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
	case "SH", "SZ", "sh", "sz":
		code = item[1] + item[0]
	case "HK", "US", "us", "hk":
		code = item[0]
	default:
		return "", errors.New("代码格式不正确")
	}
	return code, nil
}

// GetMinuteChart 获取分时行情
func GetMinuteChart(code string) bson.M {
	// 获取cid
	items := GetStockList([]string{code})[0]
	cid := int(items["cid"].(float64))
	symbol := strings.Split(code, ".")[0]

	url := "https://push2.eastmoney.com/api/qt/stock/details/get?fields1=f1,f4&fields2=f51,f52,f53,f55&pos=-50000&secid="
	url += strconv.Itoa(cid) + "." + symbol
	body, err := common.NewGetRequest(url).Do()
	if err != nil {
		return bson.M{}
	}
	str := json.Get(body, "data", "details").ToString()

	var info []string
	_ = json.Unmarshal([]byte(str), &info)

	// 初始化
	results := make([]minute, 0)
	var amount, sumNet, sumMainNet, price float64

	last := json.Get(body, "data", "prePrice").ToFloat64()
	lastTime := ""
	var p *minute
	// 根据每分钟迭代
	for _, str = range info {
		item := strings.Split(str, ",")

		// 新的一分钟
		if lastTime != item[0][0:5] {
			p = &minute{
				Time:    item[0][0:5],
				Price:   last,
				Vol:     0,
				Amount:  0,
				Net:     0,
				MainNet: 0,
			}
			lastTime = item[0][0:5]
			results = append(results, *p)
		}
		price, _ = strconv.ParseFloat(item[1], 2)
		last = price
		vol, _ := strconv.ParseFloat(item[2], 2)
		p.Vol += vol

		amount = price * vol
		p.Amount += amount

		//累加 1主动卖 2主动买
		if item[3] == "2" {
			sumNet += amount
			// 大单
			if amount >= 100000 {
				sumMainNet += amount
			}
		} else if item[3] == "1" {
			sumNet += amount * -1
			// 大单
			if amount >= 100000 {
				sumMainNet += amount * -1
			}
		}
		p.Net = sumNet
		p.MainNet = sumMainNet
	}
	return bson.M{"chart": results, "items": items}
}
