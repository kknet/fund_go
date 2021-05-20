package stock

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	URL = "https://push2ex.eastmoney.com/getStockFenShi?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wzfscj&pageindex=0&sort=1&ft=1"
)

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary
var ctx = context.Background()

type SourceData struct {
	Data struct {
		Data []struct {
			Time    int     `json:"t"`
			Price   float32 `json:"p"`
			Vol     float32 `json:"v"`
			Type    int     `json:"bs"`
			Amount  float32
			Zhudong float32
		} `json:"data"`
	} `json:"data"`
}

// CListOpt
// StockList Options
type CListOpt struct {
	Codes      []string //代码列表
	MarketType string   // 市场类型
	Sorted     bool     //排序
	SortName   string
	Search     string //搜索
	Size       int    //分页
	Page       int
}

// GetDetailData 获取单只股票图表信息
func GetDetailData(code string) interface{} {
	//最后一位
	var market = "1"
	if code[len(code)-1] == 'Z' {
		market = "0"
	}
	url := URL + "&code=" + code[0:6] + "&market=" + market
	res, err := http.Get(url)
	// 捕获异常
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	// json解析数据
	info := &SourceData{}
	err = json.Unmarshal(body, info)

	//实时分笔数据（12条）
	fenbi := make([]bson.M, 12)
	length := len(info.Data.Data)

	for i := range fenbi {
		p := &info.Data.Data[length-12+i]
		fenbi[i] = bson.M{"time": p.Time, "price": p.Price / 1000, "vol": p.Vol, "type": p.Type}
	}

	allLength := 0
	// 数据处理
	for i := 1; i < len(info.Data.Data); i++ {
		// 定义指针
		p := &info.Data.Data[i]
		pLast := &info.Data.Data[i-1]

		p.Price /= 1000
		p.Time /= 100
		p.Amount = p.Price * p.Vol //成交额
		p.Zhudong = p.Amount       //主动资金

		if p.Type == 1 {
			p.Zhudong *= -1 //主动卖盘 * -1
		} else if p.Type == 4 {
			p.Zhudong *= 0 // 中性盘 * 0
		}
		//主动资金累加
		p.Zhudong += pLast.Zhudong
		//分钟内成交量累加
		if pLast.Time == p.Time {
			p.Vol += pLast.Vol
			p.Amount += pLast.Amount
		}
		if p.Time <= 930 {
			continue
		}
		// 计算最终数组长度
		if p.Time != pLast.Time {
			allLength++
		}
	}
	allLength++
	// 创建数组
	times := make([]int, allLength)
	price := make([]float32, allLength)
	vol := make([]float32, allLength)
	amount := make([]float32, allLength)
	avg := make([]float32, allLength)
	zhudong := make([]float32, allLength)
	//添加数据
	index := 0
	for i := 1; i < len(info.Data.Data); i++ {
		// 定义指针
		p := &info.Data.Data[i]
		pLast := &info.Data.Data[i-1]
		if p.Time <= 930 {
			continue
		}
		// 添加分钟内最后一条数据
		if p.Time != pLast.Time {
			times[index] = pLast.Time
			price[index] = pLast.Price
			vol[index] = pLast.Vol
			amount[index] = pLast.Amount
			zhudong[index] = pLast.Zhudong
			index++
		}
	}
	//添加最后一条数据
	temp := len(info.Data.Data) - 1
	p := &info.Data.Data[temp]
	times[index] = p.Time
	price[index] = p.Price
	vol[index] = p.Vol
	zhudong[index] = p.Zhudong

	var sumAmount float32
	var sumVol float32
	var i int
	for i = range vol {
		sumAmount += amount[i]
		sumVol += vol[i]
		avg[i] = sumAmount / sumVol
	}
	avg = append(avg, avg[i])

	// 字典类型
	mapData := bson.M{
		"chart": bson.M{
			"times": times, "price": price, "vol": vol, "avg": avg, "zhudong": zhudong,
		},
		"ticks": fenbi,
		"items": GetStockList(CListOpt{Codes: []string{code}}),
	}
	return mapData
}

// GetStockList 获取多只股票信息
func GetStockList(opt CListOpt) []bson.M {
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
func GetMinuteChart(code string) []bson.M {
	// 从雪球获取数据
	url := "https://stock.xueqiu.com/v5/stock/chart/minute.json?period=1d&symbol=" + code
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	request.Header.Add("cookie", "device_id=24700f9f1986800ab4fcc880530dd0ed; s=dk11bk7hr3; cookiesu=301620717341066; remember=1; xq_a_token=986e48f0d816bca49abf998420bd5f7a9df0c506; xqat=986e48f0d816bca49abf998420bd5f7a9df0c506; xq_id_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1aWQiOjM2MTE0MDQxNTUsImlzcyI6InVjIiwiZXhwIjoxNjIzNTA1OTMxLCJjdG0iOjE2MjA5MTM5MzE5OTMsImNpZCI6ImQ5ZDBuNEFadXAifQ.W2LqlQRexNO2VXk0BV91L_uvm9ssWyTYJho51017TI-IRLnkKu6sB35_ZOR1z4XsvnRMSmNlTRDvMKEiapXY4VUu66ySZv3OIzHWaPkxIxBK4cSnL7CFr6CTX0OMAuHuZNHnR-1OJBA5-bPafC47AW0SvJQEs_IBCB83GZK3M859ipuVp_Hn8S0qXbg9v91U-nf4qJXQ4GOT9pjBFQ08u_KagtmfcOfoec23_ejXfrQt_X0F6EKO_w5_LwY0iQmEhE7kM8MiQjOyF6zLOY2JBbnyEkULY4uce5IClP7snpHJp1icydWQsV-eJjlGW9EmVvcDxpIiDvXVG7zfVfjtog; xq_r_token=7193b60d61d2e4db36b3dd1a465837dff68f6400; xq_is_login=1; u=3611404155; bid=3c6bb14598fe9ac45474be34ecb46d45_komyayku; Hm_lvt_1db88642e346389874251b5a1eded6e3=1621264280,1621305028,1621311654,1621324081")

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
	for i, x := range temp {
		results[i] = bson.M{
			"price": x["current"], "vol": x["volume"], "avg": x["avg_price"], "timestamp": x["timestamp"],
			"amount": x["amount"],
			"vol_compare": bson.M{
				"now":  x["volume_compare"].(map[string]interface{})["volume_sum"],
				"last": x["volume_compare"].(map[string]interface{})["volume_sum_last"],
			},
			"money_flow": x["capital"],
		}
	}
	return results
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

// FormatStock 股票代码格式化
func FormatStock(input string) string {
	return input
}

// GetRealtimeTicks 获取实时分笔成交
func GetRealtimeTicks() {

}
