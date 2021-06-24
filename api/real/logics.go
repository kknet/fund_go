package real

import (
	"errors"
	"fund_go2/common"
	"fund_go2/download"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
)

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// GetStockList 获取多只股票信息
func GetStockList(codes []string) []map[string]interface{} {
	var data []map[string]interface{}
	for _, df := range download.AllStock {
		data = append(data, df.Filter(dataframe.F{Colname: "code", Comparator: series.In, Comparando: codes}).Maps()...)
	}
	// 排序整理
	var results []map[string]interface{}
	for _, c := range codes {
		for _, item := range data {
			if item["code"].(string) == c {
				// 添加行业数据
				results = append(results, item)
				break
			}
		}
	}
	return results
}

// AddSimpleMinute 添加简略分时行情
func AddSimpleMinute(items map[string]interface{}) {
	var info []string
	url := "https://push2.eastmoney.com/api/qt/stock/trends2/get?fields1=f1,f5,f8,f10,f11&fields2=f53&iscr=0&secid="

	body, _ := common.NewGetRequest(url + items["cid"].(string)).Do()
	total := json.Get(body, "data", "trendsTotal").ToInt()
	preClose := json.Get(body, "data", "preClose").ToFloat32()
	timestamp := json.Get(body, "data", "time").ToInt()
	json.Get(body, "data", "trends").ToVal(&info)

	// 间隔
	space := 3
	results := make([]float64, 0)
	times := make([]string, 0)

	for i := 0; i < len(info); i += space {
		item := strings.Split(info[i], ",")
		data, _ := strconv.ParseFloat(item[1], 8)
		results = append(results, data)
		times = append(times, item[0])
	}
	times = append(times, "now")
	results = append(results, items["price"].(float64))

	items["chart"] = bson.M{
		"time": times, "price": results,
		"total": total / space, "close": preClose, "timestamp": timestamp,
	}
}

// Add60day 添加60日行情
func Add60day(items map[string]interface{}) {
	url := "https://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f1,f6&fields2=f51,f53&klt=101&fqt=0&end=20500101&lmt=60&secid="
	body, _ := common.NewGetRequest(url + items["cid"].(string)).Do()

	var info []string
	preClose := json.Get(body, "data", "preKPrice").ToFloat32()
	json.Get(body, "data", "klines").ToVal(&info)

	df := dataframe.ReadCSV(strings.NewReader("time,price\n" + strings.Join(info, "\n")))

	items["chart"] = bson.M{
		"time": df.Col("time").Records(), "price": df.Col("price").Float(),
		"total": 60, "close": preClose,
	}
}

// GetMinuteData 获取分时行情
func GetMinuteData(code string) interface{} {
	cid := GetStockList([]string{code})
	if len(cid) == 0 {
		return errors.New("该代码不存在")
	}
	url := "https://push2.eastmoney.com/api/qt/stock/trends2/get?fields1=f1,f5,f8&fields2=f51,f53,f56,f57,f58&iscr=0&secid="
	body, _ := common.NewGetRequest(url + cid[0]["cid"].(string)).Do()

	var info []string
	json.Get(body, "data", "trends").ToVal(&info)
	preClose := json.Get(body, "data", "preClose").ToFloat64()
	df := dataframe.ReadCSV(strings.NewReader("time,price,vol,amount,avg\n" + strings.Join(info, "\n")))

	// 添加color
	color := make([]int, df.Nrow())
	for index, i := range df.Col("price").Float() {
		if i >= preClose {
			color[index] = 1
		} else {
			color[index] = 0
		}
		preClose = i
	}
	return bson.M{
		"time": df.Col("time").Records(), "price": df.Col("price").Float(),
		"vol": df.Col("vol").Float(), "amount": df.Col("amount").Float(),
		"avg": df.Col("avg").Float(), "color": color,
	}
}

// Search 搜索股票
func search(input string) []map[string]interface{} {
	var results []map[string]interface{}

	indexes := []string{"code", "name", "pct_chg", "price", "type", "marketType", "amount"}
	// 优先展示CN, HK、US按成交额自由排序
	for _, mkt := range []string{"CN", "US", "CNIndex"} {
		df := download.AllStock[mkt].Select(indexes)
		if mkt == "CN" {
			df = df.RBind(download.AllStock["HK"].Select(indexes)).Arrange(dataframe.RevSort("amount"))
		}

		for i, item := range df.Select([]string{"code", "name"}).Records() {
			//转化为大写
			input = strings.ToUpper(input)
			str := strings.ToUpper(item[0] + " " + item[1])

			if strings.Contains(str, input) {
				if i >= 1 {
					results = append(results, df.Subset(i-1).Maps()...)
					if len(results) > 10 {
						return results
					}
				}
			}
		}
	}
	return results
}

// getRank 全市场排行
func getRank(opt *common.RankOpt) []map[string]interface{} {
	indexes := make([]int, 20)

	for i := 0; i < 20; i++ {
		indexes[i] = (opt.Page-1)*20 + i
	}
	order := dataframe.RevSort(opt.SortName)
	if opt.Sorted == true {
		order = dataframe.Sort(opt.SortName)
	}
	data := download.AllStock[opt.MarketType].Arrange(order).Subset(indexes)

	return data.Maps()
}

// PanKou 获取五档挂单明细
func PanKou(code string) bson.M {
	// 格式化代码为雪球格式
	code, err := formatStock(code)
	if err != nil {
		return bson.M{"msg": "代码格式错误"}
	}
	url := "https://stock.xueqiu.com/v5/stock/realtime/pankou.json?&symbol=" + code
	body, err := common.NewGetRequest(url).Do()
	if err != nil {
		panic(err)
	}
	str := json.Get(body, "data").ToString()
	// json解析
	var data bson.M
	_ = json.Unmarshal([]byte(str), &data)
	return data
}

// GetRealtimeTicks 获取最近分笔成交
func GetRealtimeTicks(code string) (interface{}, error) {
	cid := GetStockList([]string{code})
	if len(cid) == 0 {
		return nil, errors.New("改代码不存在")
	}

	url := "https://push2.eastmoney.com/api/qt/stock/details/get?fields1=f1&fields2=f51,f52,f53,f55&pos=-50&secid="
	body, err := common.NewGetRequest(url + cid[0]["cid"].(string)).Do()
	if err != nil {
		return nil, errors.New("请求发生错误")
	}
	var info []string
	json.Get(body, "data", "details").ToVal(&info)

	df := dataframe.ReadCSV(strings.NewReader("time,price,vol,type\n" + strings.Join(info, "\n")))

	// 更改type
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
	return df.Maps(), nil
}

// formatStock 股票代码格式化为雪球代码
func formatStock(input string) (string, error) {
	if strings.Contains(input, ".") {
		item := strings.Split(input, ".")

		switch item[1] {
		case "SH", "SZ":
			return item[1] + item[0], nil
		case "HK", "US":
			return item[0], nil
		}
	}
	return "", errors.New("代码格式不正确")
}

// getNumbers 获取涨跌分布
func getNumbers(marketType string) bson.M {
	df := download.AllStock[marketType]

	label := []string{"<10", "<7", "7-5", "5-3", "3-0", "0", "0-3", "3-5", "5-7", ">7", ">10"}
	value := []int{
		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Less, Comparando: -10}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.GreaterEq, Comparando: -20}).
			Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Less, Comparando: -7}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Less, Comparando: -5}).
			Filter(dataframe.F{Colname: "pct_chg", Comparator: series.GreaterEq, Comparando: -7}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Less, Comparando: -3}).
			Filter(dataframe.F{Colname: "pct_chg", Comparator: series.GreaterEq, Comparando: -5}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Less, Comparando: 0}).
			Filter(dataframe.F{Colname: "pct_chg", Comparator: series.GreaterEq, Comparando: -3}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Eq, Comparando: 0}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Greater, Comparando: 0}).
			Filter(dataframe.F{Colname: "pct_chg", Comparator: series.LessEq, Comparando: 3}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Greater, Comparando: 3}).
			Filter(dataframe.F{Colname: "pct_chg", Comparator: series.LessEq, Comparando: 5}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Greater, Comparando: 5}).
			Filter(dataframe.F{Colname: "pct_chg", Comparator: series.LessEq, Comparando: 7}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Greater, Comparando: 7}).Nrow(),

		df.Filter(dataframe.F{Colname: "pct_chg", Comparator: series.Eq, Comparando: 10}).Nrow(),
	}

	if marketType == "CN" {
		label[0] = "跌停"
		value[0] = df.Filter(dataframe.F{Colname: "wb", Comparator: series.Eq, Comparando: -100}).Nrow()
		label[10] = "涨停"
		value[10] = df.Filter(dataframe.F{Colname: "wb", Comparator: series.Eq, Comparando: 100}).Nrow()
	}
	return bson.M{"label": label, "value": value}
}

// GetNorthFlow 北向资金流向
func GetNorthFlow() interface{} {
	url := "https://push2.eastmoney.com/api/qt/kamt.rtmin/get?fields1=f1,f3&fields2=f52,f54,f56"
	body, _ := common.NewGetRequest(url).Do()

	var str []string
	json.Get(body, "data", "s2n").ToVal(&str)

	df := dataframe.ReadCSV(strings.NewReader("hgt,sgt,all\n" + strings.Join(str, "\n")))
	return bson.M{
		"hgt": df.Col("hgt").Float(),
		"sgt": df.Col("sgt").Float(),
		"all": df.Col("all").Float(),
	}
}

// GetMainNetFlow 获取大盘主力资金流向
func GetMainNetFlow() interface{} {
	url := "https://push2.eastmoney.com/api/qt/stock/fflow/kline/get?lmt=0&klt=1&fields1=f1,f2,f3&fields2=f51,f52,f53,f54,f55,f56&secid=1.000001&secid2=0.399001"
	body, _ := common.NewGetRequest(url).Do()

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
