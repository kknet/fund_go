package real

import (
	"context"
	"fund_go2/common"
	"fund_go2/download"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	SimpleMinuteUrl = "https://push2.eastmoney.com/api/qt/stock/trends2/get?fields1=f1,f5,f8,f10,f11&fields2=f53&iscr=0&secid="
	PanKouUrl       = "https://stock.xueqiu.com/v5/stock/realtime/pankou.json?&symbol="
	TicksUrl        = "https://push2.eastmoney.com/api/qt/stock/details/get?fields1=f1&fields2=f51,f52,f53,f55"
)

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary
var ctx = context.Background()

// query options
var basicOpt = bson.M{
	"_id": 0, "cid": 1, "code": 1, "name": 1, "type": 1, "marketType": 1,
	"close": 1, "price": 1, "pct_chg": 1, "amount": 1, "mc": 1, "tr": 1,
	"net": 1, "main_net": 1, "roe": 1, "income_yoy": 1, "revenue_yoy": 1,
}
var allOpt = bson.M{"adj_factor": 0, "_id": 0, "sw_code": 0}

// GetStockList 获取多只股票信息
func GetStockList(codes []string) []bson.M {

	// 多只股票
	if len(codes) > 1 {
		var results []bson.M
		var data []bson.M
		_ = download.RealColl.Find(ctx, bson.M{"_id": bson.M{"$in": codes}}).Select(basicOpt).All(&data)

		// 排序
		for _, c := range codes {
			for _, item := range data {
				if c == item["code"] {
					results = append(results, item)
					break
				}
			}
		}
		return results

		// 单只股票
	} else if len(codes) == 1 {
		var data bson.M
		_ = download.RealColl.Find(ctx, bson.M{"_id": codes[0]}).Select(allOpt).One(&data)

		// 添加行业数据
		group := sync.WaitGroup{}
		group.Add(3)

		addIndustry := func(ids string) {
			var info bson.M
			if data[ids] != nil {
				_ = download.RealColl.Find(ctx, bson.M{"name": data[ids], "type": ids}).
					Select(bson.M{"_id": 0, "members": 0, "marketType": 0, "type": 0}).One(&info)

				data[ids] = info
			}
			group.Done()
		}
		go addIndustry("industry")
		go addIndustry("sw")
		go addIndustry("area")

		// 添加市场状态
		marketType := data["marketType"].(string)
		data["status"] = download.Status[marketType]
		data["status_name"] = download.StatusName[marketType]

		group.Wait()
		return []bson.M{data}

	} else {
		return []bson.M{}
	}
}

// AddSimpleMinute 添加简略分时行情
func AddSimpleMinute(items bson.M) {
	cid, ok := items["cid"].(string)
	if !ok {
		items["chart"] = bson.M{"total": 0, "price": nil, "close": nil}
		return
	}
	var info []string
	res, _ := http.Get(SimpleMinuteUrl + cid)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	total := json.Get(body, "data", "trendsTotal").ToInt()
	preClose := json.Get(body, "data", "preClose").ToFloat32()
	json.Get(body, "data", "trends").ToVal(&info)

	// 间隔
	space := 3
	results := make([]float64, 0)

	for i := 0; i < len(info); i += space {
		item := strings.Split(info[i], ",")
		data, _ := strconv.ParseFloat(item[1], 8)
		results = append(results, data)
	}
	results = append(results, items["price"].(float64))

	items["chart"] = bson.M{
		"total": total / space, "price": results, "close": preClose,
	}
}

// Add60day 添加60日行情
func Add60day(items bson.M) {
	url := "https://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1,f6&fields2=f51,f53&klt=101&fqt=0&end=20500101&lmt=60&secid="
	res, _ := http.Get(url + items["cid"].(string))
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	var info []string
	preClose := json.Get(body, "data", "preKPrice").ToFloat32()
	json.Get(body, "data", "klines").ToVal(&info)

	df := dataframe.ReadCSV(strings.NewReader("time,price\n" + strings.Join(info, "\n")))

	items["chart"] = bson.M{
		"time": df.Col("time").Records(), "price": df.Col("price").Float(), "close": preClose,
	}
}

// Search 搜索股票
func search(input string) []bson.M {
	var results []bson.M

	// 模糊查询
	char := strings.Split(input, "")
	matchStr := strings.Join(char, ".*")

	// 先搜索股票
	for _, marketType := range []string{"CN", "HK", "US"} {
		var temp []bson.M

		_ = download.RealColl.Find(ctx, bson.M{
			"marketType": marketType, "type": "stock", "$or": bson.A{
				// 正则匹配 不区分大小写
				bson.M{"_id": bson.M{"$regex": matchStr, "$options": "i"}},
				bson.M{"name": bson.M{"$regex": matchStr, "$options": "i"}},
			},
		}).Sort("-amount").Select(basicOpt).Limit(10).All(&temp)
		results = append(results, temp...)

		if len(results) >= 10 {
			return results[0:10]
		}
	}

	// 指数
	var temp []bson.M
	_ = download.RealColl.Find(ctx, bson.M{
		"type": bson.M{"$ne": "stock"}, "$or": bson.A{
			// 正则匹配 不区分大小写
			bson.M{"_id": bson.M{"$regex": matchStr, "$options": "i"}},
			bson.M{"name": bson.M{"$regex": matchStr, "$options": "i"}},
		},
	}).Sort("-amount").Select(basicOpt).Limit(10).All(&temp)
	results = append(results, temp...)

	if len(results) >= 10 {
		return results[0:10]
	}
	return results
}

// getRank 市场排行
func getRank(opt *common.RankOpt) []bson.M {
	var results []bson.M

	if !opt.Sorted {
		opt.SortName = "-" + opt.SortName
	}
	_ = download.RealColl.Find(ctx, bson.M{
		"marketType": opt.MarketType,
		"vol":        bson.M{"$gt": 0},
		"type":       "stock",
	}).Sort(opt.SortName).Select(basicOpt).Skip(15 * (opt.Page - 1)).Limit(15).All(&results)
	return results
}

// GetRealTicks 获取五档挂单明细、分笔成交
func GetRealTicks(code string, count int) bson.M {
	cid := GetStockList([]string{code})
	if len(cid) <= 0 {
		return nil
	}
	result := bson.M{}
	group := sync.WaitGroup{}
	group.Add(2)

	go func() {
		// CN股票才有盘口数据
		if cid[0]["marketType"] == "CN" {
			items := strings.Split(cid[0]["code"].(string), ".")
			res, _ := http.Get(PanKouUrl + items[1] + items[0])
			body, _ := ioutil.ReadAll(res.Body)
			defer res.Body.Close()

			var data bson.M
			json.Get(body, "data").ToVal(&data)
			result["pankou"] = data
		} else {
			result["pankou"] = nil
		}
		group.Done()
	}()
	go func() {
		url := TicksUrl + "&pos=-" + strconv.Itoa(count) + "&secid=" + cid[0]["cid"].(string)
		res, _ := http.Get(url)
		body, _ := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		var info []string
		json.Get(body, "data", "details").ToVal(&info)

		df := dataframe.ReadCSV(strings.NewReader("time,price,vol,type\n" + strings.Join(info, "\n")))
		side := df.Col("type").Map(func(data series.Element) series.Element {
			switch data.Float() {
			case 4:
				data.Set(0)
			case 1:
				data.Set(-1)
			case 2:
				data.Set(1)
			}
			return data
		})
		df = df.Mutate(side)
		result["ticks"] = df.Maps()
		group.Done()
	}()
	group.Wait()
	return result
}

// getNumbers 获取涨跌分布
func getNumbers(marketType string) bson.M {
	label := []string{"跌停", "<7", "7-5", "5-3", "3-0", "0", "0-3", "3-5", "5-7", ">7", "涨停"}
	num := make([]int64, 11)

	match := []bson.M{
		{"wb": -100},
		{"pct_chg": bson.M{"$lt": -7}},
		{"pct_chg": bson.M{"$lt": -5, "$gte": -7}},
		{"pct_chg": bson.M{"$lt": -3, "$gte": -5}},
		{"pct_chg": bson.M{"$lt": -0, "$gte": -3}},
		{"pct_chg": bson.M{"$eq": 0}},
		{"pct_chg": bson.M{"$gt": 0, "$lte": 3}},
		{"pct_chg": bson.M{"$gt": 3, "$lte": 5}},
		{"pct_chg": bson.M{"$gt": 5, "$lte": 7}},
		{"pct_chg": bson.M{"$gt": 7}},
		{"wb": 100},
	}
	if marketType != "CN" {
		label[0] = "<10"
		label[10] = ">10"
		match[0] = bson.M{"pct_chg": bson.M{"$lt": -10}}
		match[10] = bson.M{"pct_chg": bson.M{"$gt": 10}}
	}
	for i := range match {
		match[i]["marketType"] = marketType
		match[i]["type"] = "stock"
		num[i], _ = download.RealColl.Find(ctx, match[i]).Count()
	}

	return bson.M{"label": label, "value": num}
}

// GetMainNetFlow 获取大盘主力资金流向
func GetMainNetFlow() interface{} {
	url := "https://push2.eastmoney.com/api/qt/stock/fflow/kline/get?lmt=0&klt=1&fields1=f1,f2,f3&fields2=f51,f52,f53,f54,f55,f56&secid=1.000001&secid2=0.399001"
	res, _ := http.Get(url)
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	var str []string
	json.Get(body, "data", "klines").ToVal(&str)

	df := dataframe.ReadCSV(strings.NewReader("time,main_net,small,mid,big,huge\n" + strings.Join(str, "\n")))

	return bson.M{
		"time":     df.Col("time").Records(),
		"main_net": df.Col("main_net").Float(),
		"small":    df.Col("small").Float(),
		"mid":      df.Col("mid").Float(),
		"big":      df.Col("big").Float(),
		"huge":     df.Col("huge").Float(),
	}
}

// GetIndustryMembers 获取板块成分股
func GetIndustryMembers(industryCode string) []bson.M {
	var members bson.M
	var data []bson.M
	_ = download.RealColl.Find(ctx, bson.M{"_id": industryCode}).Select(bson.M{"members": 1}).One(&members)
	_ = download.RealColl.Find(ctx, bson.M{"_id": bson.M{"$in": members["members"]}}).Select(basicOpt).All(&data)

	return data
}
