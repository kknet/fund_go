package download

import (
	"fund_go2/common"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
	"gonum.org/v1/gonum/mat"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var MyChan = getGlobalChan()

// 更新频率
const (
	MaxCount = 500
	MidCount = 10
)

// Status 市场状态：是否开市
var Status = map[string]bool{
	"CN": false, "HK": false, "US": false,
}

// StatusName 市场状态描述：盘前交易、交易中、休市中、已收盘、休市
var StatusName = map[string]string{
	"CN": "", "HK": "", "US": "",
}

// 市场参数
var fs = map[string]string{
	"CNIndex": "m:1+s:2,m:0+t:5",
	"CN":      "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",
	"HK":      "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4",
	"US":      "m:105,m:106,m:107",
}

// 低频数据（开盘时更新）
var lowName = map[string]string{
	"f13": "cid", "f14": "name", "f18": "close",
	"f38": "total_share", "f39": "float_share",
	"f37": "roe", "f40": "revenue", "f41": "revenue_yoy", "f45": "income", "f46": "income_yoy",
}

// 中频数据（约每分钟更新）
var basicName = map[string]string{
	"f17": "open", "f23": "pb", "f115": "pe_ttm",
	"f8": "tr", "f10": "vr", "f20": "mc", "f21": "fmc",
	"f267": "3day_main_net", "f164": "5day_main_net", "f174": "10day_main_net",
}

// 高频数据（毫秒级更新）
var proName = map[string]string{
	"f12": "code", "f2": "price", "f15": "high", "f16": "low", "f3": "pct_chg",
	"f5": "vol", "f6": "amount", "f33": "wb", "f34": "buy", "f35": "sell",
	"f62": "main_net",
}

// 初始化全局通道
func getGlobalChan() chan string {
	var ch chan string
	var chanOnceManager sync.Once

	chanOnceManager.Do(func() {
		ch = make(chan string)
	})
	return ch
}

// 计算股票指标
func calData(df dataframe.DataFrame, marketType string) dataframe.DataFrame {

	code := df.Col("code").Records()
	// cid
	if common.InSlice("cid", df.Names()) {
		cid := df.Col("cid").Records()
		for i := range cid {
			cid[i] += "." + code[i]
		}
		df = df.Mutate(series.New(cid, series.String, "cid"))
	}

	// code
	for i := range code {
		switch marketType {
		case "CN":
			code[i] += Expression(code[i][0] == '6', ".SH", ".SZ").(string)
		case "CNIndex":
			code[i] += Expression(code[i][0] == '0', ".SH", ".SZ").(string)
		case "HKIndex", "USIndex":
			code[i] += "." + marketType[0:2]
		case "HK", "US":
			code[i] += "." + marketType
		}
	}
	df = df.Mutate(series.New(code, series.String, "code"))

	// net
	avgPrice := Cal(df.Col("amount"), "/", df.Col("vol"))
	buy := Cal(df.Col("buy"), "-", df.Col("sell"))
	net := Cal(avgPrice, "*", buy, "net")
	df = df.Mutate(net).Drop([]string{"buy", "sell"})

	return df
}

// Cal series之间运算
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
	url := "https://push2.eastmoney.com/api/qt/clist/get?po=1&fid=f20&pz=5000&np=1&fltt=2&pn=1&fs=" + fs[marketType] + "&fields="
	// HK US指数
	if marketType == "HKIndex" {
		url = "https://push2.eastmoney.com/api/qt/ulist.np/get?secids=100.HSI,100.HSCEI,124.HSCCI&fltt=2&fields="
	} else if marketType == "USIndex" {
		url = "https://push2.eastmoney.com/api/qt/ulist.np/get?secids=100.DJIA,100.SPX,100.NDX&fltt=2&fields="
	}
	var tempUrl string
	// 定时更新计数器
	var count = MaxCount
	client := &http.Client{}
	for {
		// 连接参数
		tempUrl = url + common.JoinMapKeys(proName, ",")
		if count%MidCount == 0 {
			tempUrl += "," + common.JoinMapKeys(basicName, ",")
		}
		if count%MaxCount == 0 {
			tempUrl += "," + common.JoinMapKeys(lowName, ",")
		}

		res, err := client.Get(tempUrl)
		if err != nil {
			log.Println("下载股票数据失败，3秒后重试...", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		body, _ := ioutil.ReadAll(res.Body)
		str := json.Get(body, "data", "diff").ToString()
		_ = res.Body.Close()

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
		UpdateMongo(df.Maps())

		// 更新行业数据
		if marketType == "CN" {
			if count%10 == 0 {
				go CalIndustry()
			}
		}
		count++
		MyChan <- marketType

		// 重置计数器
		if count > MaxCount {
			count = 0
		}

		for !Status[marketType[0:2]] {
			count = MaxCount
			time.Sleep(time.Millisecond * 300)
		}
		time.Sleep(time.Millisecond * 300)
	}
}

// 获取市场交易状态
func getMarketStatus() {
	url := "https://xueqiu.com/service/v5/stock/batch/quote?symbol=SH000001,HKHSI,.IXIC"
	client := &http.Client{}
	for {
		res, err := client.Get(url)
		if err != nil {
			log.Println("更新市场状态失败，3秒后重试...", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}

		body, _ := ioutil.ReadAll(res.Body)
		items := json.Get(body, "data", "items").ToString()
		_ = res.Body.Close()

		// 设置CN，HK，US市场状态
		for i := 0; i < 3; i++ {
			// 解析数据
			market := json.Get([]byte(items), i, "market", "region").ToString()
			statusName := json.Get([]byte(items), i, "market", "status").ToString()
			status := Expression(statusName == "交易中", true, false).(bool)

			Status[market] = status
			StatusName[market] = statusName
		}
		// 每秒更新
		time.Sleep(time.Second * 1)
	}
}

// GoDownload 主函数
func GoDownload() {
	go getMarketStatus()
	go getRealStock("CN")
	go getRealStock("CNIndex")
	go getRealStock("HK")
	go getRealStock("US")
	go getRealStock("HKIndex")
	go getRealStock("USIndex")
}
