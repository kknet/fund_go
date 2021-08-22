package download

import (
	"fund_go2/common"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
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

// Status 市场状态
// 盘前交易、交易中、已收盘、休市
var Status = map[string]bool{
	"CN": false, "HK": false, "US": false, "Index": false,
}
var StatusName = map[string]string{
	"CN": "", "HK": "", "US": "", "Index": "",
}

// 参数
var fs = map[string]string{
	"Index": "m:1+s:2,m:0+t:5",
	"CN":    "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",
	"HK":    "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4",
	"US":    "m:105,m:106,m:107",
}

// 低频更新
var lowName = map[string]string{
	"f13": "cid", "f14": "name", "f18": "close",
	"f38": "total_share", "f39": "float_share",
	"f37": "roe", "f40": "revenue", "f41": "revenue_yoy", "f45": "income", "f46": "income_yoy",
}

// 中频更新
var basicName = map[string]string{
	"f17": "open", "f23": "pb", "f115": "pe_ttm", "f10": "vr",
	"f267": "3day_main_net", "f164": "5day_main_net", "f174": "10day_main_net",
}

// 高频更新
var proName = map[string]string{
	"f12": "code", "f2": "price", "f15": "high", "f16": "low", "f3": "pct_chg",
	"f5": "vol", "f6": "amount", "f33": "wb", "f34": "buy", "f35": "sell",
	"f64": "huge_in", "f65": "huge_out", "f70": "big_in", "f71": "big_out",
	"f78": "main_mid", "f84": "main_small",
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
		case "Index":
			code[i] += Expression(code[i][0] == '0', ".SH", ".SZ").(string)
		case "HK", "US":
			code[i] += "." + marketType
		}
	}
	df = df.Mutate(series.New(code, series.String, "code"))

	// net
	vol := df.Col("vol").Float()
	amount := df.Col("amount").Float()
	buy := df.Col("buy").Float()
	sell := df.Col("sell").Float()
	for i := range buy {
		buy[i] = (buy[i] - sell[i]) * amount[i] / vol[i]
	}
	df = df.Mutate(series.New(buy, series.Float, "net"))

	// tr mc fmc
	if common.InSlice("total_share", df.Names()) {
		price := df.Col("price").Float()
		tr := df.Col("vol").Float()
		mc := df.Col("total_share").Float()
		fmc := df.Col("float_share").Float()

		for i := range price {
			if fmc[i] > 0 {
				tr[i] = tr[i] / fmc[i] * 10000
			}
			mc[i] = price[i] * mc[i]
			fmc[i] = price[i] * fmc[i]
		}
		df = df.Mutate(series.New(tr, series.Float, "tr"))
		df = df.Mutate(series.New(mc, series.Float, "mc"))
		df = df.Mutate(series.New(fmc, series.Float, "fmc"))
	}

	// 主力净流入
	if common.InSlice("huge_in", df.Names()) && common.InSlice(marketType, []string{"CN", "HK"}) {
		main := df.Select([]string{"huge_in", "huge_out", "big_in", "big_out"}).Rapply(func(s series.Series) series.Series {
			value := s.Float()

			mainIn := value[0] + value[2]
			mainOut := value[1] + value[3]

			return series.Floats([]float64{
				value[0] - value[1], // 超大
				value[2] - value[3], // 大
				mainIn,              // 流入
				mainOut,             // 流出
				mainIn - mainOut,    // 净流入
			})
		})
		_ = main.SetNames("main_huge", "main_big", "main_in", "main_out", "main_net")
		df = df.CBind(main).Drop([]string{"huge_in", "huge_out", "big_in", "big_out"})
	}

	return df
}

// 下载数据
func getRealStock(marketType string) {
	url := "https://push2.eastmoney.com/api/qt/clist/get?po=1&fid=f20&pz=4600&np=1&fltt=2&pn=1&fs=" + fs[marketType] + "&fields="
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

		if marketType == "CN" {
			// 间隔更新行业数据
			if count%10 == 0 {
				go CalIndustry()
			}
		}
		count++
		MyChan <- marketType

		if count >= MaxCount {
			count = 0
		}
		for !Status[marketType] {
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
			if market == "CN" {
				Status["Index"] = status
				StatusName["Index"] = statusName
			}
		}
		// 每秒更新
		time.Sleep(time.Second * 1)
	}
}

// GoDownload 主函数
func GoDownload() {
	go getMarketStatus()
	go getRealStock("CN")
	go getRealStock("Index")
	go getRealStock("HK")
	go getRealStock("US")
}
