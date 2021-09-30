package download

import (
	"fund_go2/common"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
	"gonum.org/v1/gonum/mat"
	"log"
	"strings"
	"sync"
	"time"
)

// 更新频率
const (
	MaxCount = 500
	MidCount = 10
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	// MyChan 全局通道
	MyChan chan string

	// Status 市场状态：是否开市
	Status = sync.Map{}

	// StatusName 市场状态描述：盘前交易、交易中、休市中、已收盘、休市
	StatusName = sync.Map{}

	// 市场参数
	fs = map[string]string{
		"CNIndex": "m:1+s:2,m:0+t:5",
		"CN":      "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",
		"HK":      "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4",
		"US":      "m:105,m:106,m:107",
	}
	// 低频数据（开盘时更新）
	lowName = map[string]string{
		"f13": "cid", "f14": "name", "f18": "close",
		"f37": "roe", "f40": "revenue", "f41": "revenue_yoy", "f45": "income", "f46": "income_yoy",
	}
	// 中频数据（约每分钟更新）
	basicName = map[string]string{
		"f17": "open", "f23": "pb", "f115": "pe_ttm", "f10": "vr",
		"f38": "total_share", "f39": "float_share", "f33": "wb",
		"f267": "3day_main_net", "f164": "5day_main_net", "f174": "10day_main_net",
	}
	// 高频数据（毫秒级更新）
	proName = map[string]string{
		"f12": "_id", "f2": "price", "f15": "high", "f16": "low", "f3": "pct_chg",
		"f5": "vol", "f6": "amount", "f34": "buy", "f35": "sell", "f62": "main_net",
	}
)

func init() {
	// 初始化全局通道
	var chanOnceManager sync.Once

	chanOnceManager.Do(func() {
		MyChan = make(chan string)
	})

	// map
	Status.Store("CN", false)
	Status.Store("HK", false)
	Status.Store("US", false)
	StatusName.Store("CN", "")
	StatusName.Store("HK", "")
	StatusName.Store("US", "")
}

// 计算股票指标
func calData(df dataframe.DataFrame, marketType string) dataframe.DataFrame {
	// code 改为 _id
	code := df.Col("_id")

	// 格式化cid
	cid := df.Col("cid")
	if cid.Err == nil {
		for i, str := range cid.Records() {
			cid.Elem(i).Set(str + "." + code.Elem(i).String())
		}
		df = df.Mutate(cid)
	}

	// 格式化code
	for i, str := range code.Records() {
		switch marketType {
		case "CN":
			str += Expression(str[0] == '6', ".SH", ".SZ").(string)
		case "CNIndex":
			str += Expression(str[0] == '0', ".SH", ".SZ").(string)
		case "HK", "US":
			str += "." + marketType
		}
		code.Elem(i).Set(str)
	}
	df = df.Mutate(code)

	// net
	if df.Col("buy").Err == nil {
		avgPrice := Cal(df.Col("amount"), "/", df.Col("vol"))
		buy := Cal(df.Col("buy"), "-", df.Col("sell"))
		net := Cal(avgPrice, "*", buy, "net")
		df = df.Mutate(net).Drop([]string{"buy", "sell"})
	}

	if marketType == "CNIndex" {
		return df
	}

	// mc fmc tr
	if df.Col("total_share").Err == nil {
		df = df.Filter(dataframe.F{Colname: "total_share", Comparator: series.Greater, Comparando: 0})

		price := df.Col("price")
		tShare := df.Col("total_share")
		fShare := df.Col("float_share")

		df = df.Mutate(Cal(tShare, "*", price, "mc"))
		df = df.Mutate(Cal(fShare, "*", price, "fmc"))

		tr := Cal(df.Col("vol"), "/", fShare, "tr")

		// 如果是A股 换手率需要再乘100
		scale := Expression(marketType[0:2] == "CN", 10000.0, 100.0).(float64)
		for i, value := range tr.Float() {
			tr.Elem(i).Set(value * scale)
		}
		df = df.Mutate(tr)
	}
	return df
}

// Cal series向量运算
func Cal(s1 series.Series, operation string, s2 series.Series, name ...string) series.Series {
	v1 := mat.NewVecDense(s1.Len(), s1.Float())
	v2 := mat.NewVecDense(s2.Len(), s2.Float())

	switch operation {
	case "+":
		v1.AddVec(v1, v2)
	case "-":
		v1.SubVec(v1, v2)
	case "*":
		v1.MulElemVec(v1, v2)
	case "/":
		v1.DivElemVec(v1, v2)
	}
	if len(name) > 0 {
		return series.New(v1.RawVector().Data, series.Float, name[0])
	} else {
		return series.New(v1.RawVector().Data, series.Float, "x")
	}
}

// 下载股票数据
func getRealStock(marketType string) {
	url := "http://push2.eastmoney.com/api/qt/clist/get?po=1&fid=f6&pz=4600&np=1&fltt=2&pn=1&fs=" + fs[marketType] + "&fields="
	var tempUrl string
	// 定时更新计数器
	var count = MaxCount
	for {
		// 连接参数
		tempUrl = url + common.JoinMapKeys(proName, ",")
		if count%MidCount == 0 {
			tempUrl += "," + common.JoinMapKeys(basicName, ",")
		}
		if count%MaxCount == 0 {
			tempUrl += "," + common.JoinMapKeys(lowName, ",")
		}

		body, err := common.GetAndRead(tempUrl)
		if err != nil {
			log.Println("下载股票数据失败，3秒后重试...", err)
			time.Sleep(time.Second * 3)
			continue
		}
		str := json.Get(body, "data", "diff").ToString()

		df := dataframe.ReadJSON(strings.NewReader(str), dataframe.WithTypes(map[string]series.Type{
			"f12": series.String, "f13": series.String,
		}))

		// 重命名
		for _, col := range df.Names() {
			newName, ok := proName[col]
			if !ok {
				newName, ok = basicName[col]
				if !ok {
					newName = lowName[col]
				}
			}
			df = df.Rename(newName, col)
		}

		df = calData(df, marketType)
		updateMongo(df.Maps())

		// 更新行业数据
		if count%10 == 0 && marketType == "CN" {
			go calIndustry()
		}

		// 更新计数器
		count++
		MyChan <- marketType
		// 重置计数器
		if count > MaxCount {
			count = 0
		}

		status, _ := Status.Load(marketType[0:2])
		for !status.(bool) {
			count = MaxCount
			time.Sleep(time.Millisecond * 300)
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// 获取市场交易状态
func getMarketStatus() {
	url := "https://xueqiu.com/service/v5/stock/batch/quote?symbol=SH000001,HKHSI,.IXIC"
	for {
		body, err := common.GetAndRead(url)
		if err != nil {
			log.Println("更新市场状态失败，3秒后重试...", err)
			time.Sleep(time.Second * 3)
			continue
		}
		items := json.Get(body, "data", "items")

		// 设置CN，HK，US市场状态
		for i := 0; i < 3; i++ {
			// 市场类型（地区）
			market := items.Get(i, "market", "region").ToString()
			// 状态名称
			statusName := items.Get(i, "market", "status").ToString()
			// 状态
			Status.Store(market, Expression(statusName == "交易中", true, false))
			StatusName.Store(market, statusName)
		}
		time.Sleep(time.Second * 3)
	}
}

// GoDownload 主函数
func GoDownload() {
	go getMarketStatus()
	go getRealStock("CN")
	go getRealStock("CNIndex")
	go getRealStock("HK")
	go getRealStock("US")
}
