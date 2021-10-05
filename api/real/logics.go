package real

import (
	"context"
	"fmt"
	"fund_go2/common"
	"fund_go2/download"
	"fund_go2/env"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"sync"
)

// 东方财富url
const (
	SimpleMinuteUrl = "http://push2.eastmoney.com/api/qt/stock/trends2/get?fields1=f10&fields2=f53&iscr=0&secid="
	Day60Url        = "http://push2his.eastmoney.com/api/qt/stock/kline/get?fields1=f6&fields2=f53&klt=101&fqt=0&end=20500101&lmt=60&secid="
	PanKouUrl       = "http://push2.eastmoney.com/api/qt/stock/get?fltt=2&fields=f58,f530,f135,f136,f137,f138,f139,f141,f142,f144,f145,f147,f148,f140,f143,f146,f149&secid="
	TicksUrl        = "http://push2.eastmoney.com/api/qt/stock/details/get?fields1=f1&fields2=f51,f52,f53,f55"
	MoneyFlowUrl    = "http://push2.eastmoney.com/api/qt/stock/fflow/kline/get?lmt=0&klt=1&fields1=f1&fields2=f53,f54,f55,f56&secid="
)

var (
	// jsoniter
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	ctx  = context.Background()
	// 文档
	realColl *qmgo.Collection
	// 访问热度
	hotDB *redis.Client
	// 流量控制
	limitDB *redis.Client
	// 列表详细数据
	basicOpt = bson.M{
		"cid": 1, "name": 1, "type": 1, "marketType": 1, "close": 1,
		"price": 1, "pct_chg": 1, "amount": 1, "mc": 1, "tr": 1,
		"net": 1, "main_net": 1, "roe": 1, "income_yoy": 1, "revenue_yoy": 1,
	}
	// 列表简略数据 (websocket更新时使用)
	simpleOpt = bson.M{
		"price": 1, "pct_chg": 1, "vol": 1, "amount": 1, "net": 1, "main_net": 1,
	}
	// 搜索
	searchOpt = bson.M{
		"name": 1, "type": 1, "marketType": 1, "price": 1, "pct_chg": 1, "amount": 1,
	}
)

func init() {
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: "mongodb://" + env.MongoHost})
	if err != nil {
		panic(err)
	}
	realColl = client.Database("stock").Collection("realStock")

	hotDB = redis.NewClient(&redis.Options{
		Addr: env.RedisHost,
		DB:   1,
	})
	limitDB = redis.NewClient(&redis.Options{
		Addr: env.RedisHost,
		DB:   2,
	})
}

// GetStock 获取单只股票
// detail: 获取详细信息（所有板块）、市场状态
func GetStock(code string, detail ...bool) bson.M {
	var data bson.M
	_ = realColl.Find(ctx, bson.M{"_id": code}).Select(bson.M{"adj_factor": 0}).One(&data)

	if len(data) <= 0 {
		return nil
	}

	if len(detail) > 0 {
		// 股票板块
		if data["type"] == "stock" {
			var bk []bson.M

			_ = realColl.Find(ctx, bson.M{
				"_id": bson.M{"$in": data["bk"]},
				// 排序是为了控制板块顺序，industry最先，concept在中间，area最后
			}).Select(bson.M{"name": 1, "type": 1, "pct_chg": 1}).Sort("-type").All(&bk)
			data["bk"] = bk
		}
		// 添加市场状态
		market := data["marketType"].(string)
		data["status"] = download.MarketStatus[market].Status
		data["status_name"] = download.MarketStatus[market].StatusName
	}
	return data
}

// GetStockList 获取多只股票信息
// simple: 只获取简略信息
func GetStockList(codes []string, simple ...bool) []bson.M {
	results := make([]bson.M, 0)
	data := make([]bson.M, 0)

	if len(simple) > 0 {
		_ = realColl.Find(ctx, bson.M{"_id": bson.M{"$in": codes}}).Select(simpleOpt).All(&data)
	} else {
		_ = realColl.Find(ctx, bson.M{"_id": bson.M{"$in": codes}}).Select(basicOpt).All(&data)
	}

	// 排序
	for _, c := range codes {
		for _, item := range data {
			if c == item["_id"] {
				results = append(results, item)
				break
			}
		}
	}
	return results
}

// 获取同花顺行业分时行情
func getIndustryMinute(code string) interface{} {
	symbol := strings.Split(code, ".")[0]
	body, err := common.GetThsAndRead("http://d.10jqka.com.cn/v6/time/bk_" + symbol + "/last.js")
	if err != nil {
		return err
	}
	// 去掉最外层的括号
	str := strings.Split(string(body), "(")[1]
	str = str[:len(str)-1]

	info := json.Get([]byte(str), "bk_"+symbol, "data").ToString()
	info = strings.Join(strings.Split(info, ";"), "\n")
	df := dataframe.ReadCSV(strings.NewReader("time,price,amount,vol\n" + info))

	return bson.M{
		"price":  df.Col("price").Float(),
		"vol":    df.Col("vol").Float(),
		"amount": df.Col("amount").Float(),
	}
}

// 添加简略分时行情
func addSimpleMinute(items bson.M) {
	cid, ok := items["cid"].(string)
	if !ok {
		getThsSimpleMinute(items)
		return
	}
	body, err := common.GetAndRead(SimpleMinuteUrl + cid)
	if err != nil {
		return
	}

	data := json.Get(body, "data")
	total := data.Get("trendsTotal").ToInt()
	data = data.Get("trends")

	results := make([]float64, 0)

	for i := 0; i < data.Size(); i += 5 {
		item := strings.Split(data.Get(i).ToString(), ",")
		data, _ := strconv.ParseFloat(item[1], 8)
		results = append(results, data)
	}

	items["chart"] = bson.M{
		"total": total / 5, "price": results, "close": items["close"], "type": "price",
	}
}

// 获取同花顺简略分时行情
func getThsSimpleMinute(items bson.M) {
	code := items["_id"].(string)
	symbol := strings.Split(code, ".")
	code = "bk_" + symbol[0]

	body, err := common.GetThsAndRead("http://d.10jqka.com.cn/v6/time/" + code + "/last.js")
	if err != nil {
		return
	}
	// 去掉最外层的括号
	strs := strings.Split(string(body), "(")
	if len(strs) < 2 {
		return
	}
	// 取第二条 并去掉末尾的右括号
	str := strs[1]
	bytesData := common.Str2bytes(str[:len(str)-1])

	source := json.Get(bytesData, code, "data").ToString()
	total := json.Get(bytesData, code, "dotsCount").ToInt()
	preClose := json.Get(bytesData, code, "pre").ToFloat64()

	info := strings.Split(source, ";")

	results := make([]float64, 0)

	for i := 0; i < len(info); i += 5 {
		item := strings.Split(info[i], ",")
		data, _ := strconv.ParseFloat(item[1], 8)
		results = append(results, data)
	}
	items["chart"] = bson.M{
		"total": total / 5, "price": results, "close": preClose, "type": "price",
	}
}

// 添加60日行情
func add60day(items bson.M) {
	body, err := common.GetAndRead(Day60Url + items["cid"].(string))
	if err != nil {
		return
	}

	var info []string
	preClose := json.Get(body, "data", "preKPrice").ToFloat32()
	json.Get(body, "data", "klines").ToVal(&info)

	items["chart"] = bson.M{
		"total": 60, "price": info, "close": preClose, "type": "price",
	}
}

// 添加主力资金趋势
func addMainFlow(items bson.M) {
	url := "http://push2.eastmoney.com/api/qt/stock/fflow/kline/get?lmt=0&klt=1&fields1=f1&fields2=f52&secid="
	cid, ok := items["cid"].(string)
	if !ok {
		return
	}

	body, err := common.GetAndRead(url + cid)
	if err != nil {
		return
	}

	total := 80
	var info []string
	json.Get(body, "data", "klines").ToVal(&info)

	// 间隔
	space := 3
	results := make([]float64, 0)

	for i := 0; i < len(info); i += space {
		data, _ := strconv.ParseFloat(info[i], 8)
		results = append(results, data)
	}
	items["chart"] = bson.M{
		"total": total, "price": results, "close": 0, "type": "flow",
	}
}

// 搜索股票
func search(input string) []bson.M {
	var results []bson.M

	// 模糊查询
	matchStr := strings.Replace(input, "", ".*", -1)
	matchOpt := bson.A{
		// 正则匹配 不区分大小写
		bson.M{"_id": bson.M{"$regex": matchStr, "$options": "i"}},
		bson.M{"name": bson.M{"$regex": matchStr, "$options": "i"}},
	}

	// 优先级1: 板块(area,concept,industry)
	temp := make([]bson.M, 0)
	_ = realColl.Find(ctx, bson.M{
		// 只支持名字搜索
		"name": bson.M{"$regex": matchStr, "$options": "i"},

		"type": bson.M{"$in": bson.A{"industry", "area", "concept"}},
	}).Select(searchOpt).Limit(12).All(&temp)

	results = append(results, temp...)
	if len(results) >= 12 {
		return results
	}

	// 优先级2: stock(CN > HK > US)
	temp = []bson.M{}
	_ = realColl.Find(ctx, bson.M{
		"$or": matchOpt, "type": "stock",
	}).Select(searchOpt).Sort("marketType").Sort("-amount").Limit(12).All(&temp)

	results = append(results, temp...)
	if len(results) >= 12 {
		return results
	}

	// 优先级3: index
	temp = []bson.M{}
	_ = realColl.Find(ctx, bson.M{"$or": matchOpt, "type": "index"}).Select(searchOpt).Limit(12).All(&temp)

	results = append(results, temp...)
	return results
}

// 市场排行
func getRank(opt *common.RankOpt) []bson.M {
	var results []bson.M

	var rankOpt = bson.M{
		"cid": 1, "name": 1, "marketType": 1, "type": 1, "close": 1, "price": 1, "pct_chg": 1,
	}

	// 添加指定参数
	rankOpt[opt.SortName] = 1
	if !opt.Sorted {
		opt.SortName = "-" + opt.SortName
	}

	_ = realColl.Find(ctx, bson.M{
		"marketType": opt.MarketType, "type": "stock", "mc": bson.M{"$gt": 0},
	}).Sort(opt.SortName).Select(rankOpt).Skip(20 * (opt.Page - 1)).Limit(20).All(&results)

	return results
}

// GetRealTicks 获取五档挂单明细、分笔成交
func GetRealTicks(item bson.M) (bson.M, []map[string]interface{}) {
	var ticks []map[string]interface{}
	var pankou bson.M

	cid, ok := item["cid"].(string)
	if !ok {
		return nil, nil
	}
	group := sync.WaitGroup{}
	group.Add(2)

	// 获取盘口明细
	go func() {
		body, err := common.GetAndRead(PanKouUrl + cid)
		if err != nil {
			return
		}
		var data bson.M
		json.Get(body, "data").ToVal(&data)
		pankou = data
		group.Done()
	}()
	// 获取成交明细
	go func() {
		body, err := common.GetAndRead(TicksUrl + "&pos=-40&secid=" + cid)
		if err != nil {
			return
		}
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
		ticks = df.Mutate(side).Maps()
		group.Done()
	}()
	group.Wait()
	return pankou, ticks
}

// 获取市场涨跌分布
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
		num[i], _ = realColl.Find(ctx, match[i]).Count()
	}

	return bson.M{"label": label, "value": num}
}

// 获取资金博弈走势
func getDetailMoneyFlow(code string) interface{} {
	item := GetStock(code, false)
	cid, ok := item["cid"].(string)
	if !ok {
		return nil
	}

	body, err := common.GetAndRead(MoneyFlowUrl + cid)
	if err != nil {
		return nil
	}

	var str []string
	json.Get(body, "data", "klines").ToVal(&str)
	if len(str) == 0 {
		return nil
	}

	df := dataframe.ReadCSV(strings.NewReader("small,mid,big,huge\n" + strings.Join(str, "\n")))
	return bson.M{
		"small": df.Col("small").Float(),
		"mid":   df.Col("mid").Float(),
		"big":   df.Col("big").Float(),
		"huge":  df.Col("huge").Float(),
	}
}

// 获取板块成分股
func getIndustryMembers(code string) []bson.M {
	var members bson.M
	var data []bson.M
	// 获取成分股列表
	_ = realColl.Find(ctx, bson.M{"_id": code}).Select(bson.M{"members": 1}).One(&members)
	// 获取行情
	_ = realColl.Find(ctx, bson.M{"_id": bson.M{"$in": members["members"]}}).
		Sort("-pct_chg").Limit(25).Select(basicOpt).All(&data)

	return data
}

// 查看股票页面
func viewPage(code string) interface{} {
	res, err := hotDB.ZIncrBy(ctx, "hot", 1.0, code).Result()
	if err != nil {
		return err
	}
	return res
}

// 获取市场板块简略信息
func getSimpleBK(idsName string) []bson.M {
	// sync Map
	results := sync.Map{}

	query := realColl.Find(ctx, bson.M{"type": idsName}).Select(bson.M{"name": 1, "pct_chg": 1, "领涨股": 1, "max_pct": 1, "main_net": 1})

	myFunc := func(sortName string, limit int64) {
		var temp []bson.M
		_ = query.Sort(sortName).Limit(limit).All(&temp)
		for _, i := range temp {
			results.Store(i["_id"], i)
		}
	}

	myFunc("pct_chg", 6)
	myFunc("-pct_chg", 6)
	myFunc("main_net", 4)
	myFunc("-main_net", 4)

	var data []bson.M
	results.Range(func(key, value interface{}) bool {
		data = append(data, value.(bson.M))
		return true
	})
	return data
}

// 获取热门股票数据（来源：雪球）
func getXueQiuHotStock() {
	// type: 全球10 沪深12 港股13 美股11
	body, err := common.GetAndRead("http://stock.xueqiu.com/v5/stock/hot_stock/list.json?size=100&type=10")
	fmt.Println(body, err)
}
