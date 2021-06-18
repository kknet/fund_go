package download

import (
	"fund_go2/common"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
	"log"
	"strings"
	"sync"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// AllStock 存储所有股票数据
var AllStock = map[string]dataframe.DataFrame{}

// MyChan 全局通道
var MyChan = getGlobalChan()

// 将MyChan设置为全局通道
func getGlobalChan() chan interface{} {
	var ch chan interface{}
	var chanOnceManager sync.Once

	chanOnceManager.Do(func() {
		ch = make(chan interface{})
	})
	return ch
}

// 计算指标
func calData(df dataframe.DataFrame, marketType string) dataframe.DataFrame {
	//基本信息
	basic := df.Select([]string{"cid", "code"}).Rapply(func(s series.Series) series.Series {
		// cid
		cid := s.Elem(0).String() + "." + s.Elem(1).String()
		// code
		code := s.Elem(1).String()
		switch marketType {
		case "CN":
			code += Expression(code[0] == '6', ".SH", ".SZ").(string)
		case "CNIndex":
			code += Expression(code[0] == '0', ".SH", ".SZ").(string)
		case "HK", "US":
			code += "." + marketType
		}
		return series.Strings([]string{cid, code})
	})
	_ = basic.SetNames("cid", "code")

	for _, col := range basic.Names() {
		df = df.Mutate(basic.Col(col))
	}

	// 计算指标
	indexes := []string{"price", "close", "high", "low", "vol", "total_share", "float_share", "内盘", "外盘", "amount"}
	pct := df.Select(indexes).Rapply(func(s series.Series) series.Series {
		value := s.Float()

		return series.Floats([]float64{
			(value[0]/value[1] - 1.0) * 100,             // pct_chg
			(value[2] - value[3]) / value[1] * 100,      // amp
			value[4] / value[6] * 10000,                 // tr
			value[0] * value[5],                         // mc
			value[0] * value[6],                         // fmc
			(value[8] - value[7]) * value[9] / value[4], // 均价
		})
	})
	_ = pct.SetNames("pct_chg", "amp", "tr", "mc", "fmc", "net")

	for _, col := range pct.Names() {
		df = df.Mutate(pct.Col(col))
	}

	df = df.SetCol("marketType", Expression(marketType == "CNIndex", "CN", marketType))
	df = df.SetCol("type", Expression(marketType == "CNIndex", "index", "stock"))

	//主力资金流向
	if marketType == "CN" {
		data := df.Select([]string{"main_huge", "main_big", "main_pct"}).Rapply(func(s series.Series) series.Series {
			value := s.Float()
			net := value[0] + value[1]
			amount := net / value[2] * 100
			in := (net + amount) / 2.0
			out := net - in

			return series.Floats([]float64{net, in, out})
		})
		_ = data.SetNames("main_net", "main_in", "main_out")
		df = df.CBind(data)
	}
	df = df.Drop([]string{"内盘", "外盘"})
	return df
}

// 下载数据
func getEastMoney(marketType string) {
	fs := map[string]string{
		"CNIndex": "m:1+s:2,m:0+t:5",
		"CN":      "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",
		"HK":      "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4",
		"US":      "m:105,m:106,m:107",
	}
	url := "https://push2.eastmoney.com/api/qt/clist/get?po=1&fid=f6&invt=2&pz=7500&np=1&fltt=2&pn=1&fs=" + fs[marketType] + "&fields="
	// 重命名
	name := map[string]string{
		"f2": "price", "f5": "vol", "f6": "amount", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "vr", "f13": "cid", "f14": "name", "f18": "close",
		"f23": "pb", "f34": "外盘", "f35": "内盘",
		"f22": "pct_rate", "f11": "pct5min", "f24": "pct60day", "f25": "pct_current_year",
		"f38": "total_share", "f39": "float_share", "f115": "pe_ttm",
		//"f100": "EMIds",
		"f37": "roe",
		//"f40": "营收", "f41": "营收同比", "f45": "净利润", "f46": "净利润同比",
	}
	if marketType == "CN" {
		name["f33"] = "wb"
		name["f66"] = "main_huge"
		name["f72"] = "main_big"
		name["f78"] = "main_mid"
		name["f84"] = "main_small"
		name["f184"] = "main_pct"
	}
	//连接参数
	for i := range name {
		url += i + ","
	}
	url = url[:len(url)-1]

	request := common.NewGetRequest(url)
	for {
		body, err := request.Do()
		if err != nil {
			log.Println("下载股票数据发生错误，", err.Error())
		}
		str := json.Get(body, "data", "diff").ToString()
		str = strings.Replace(str, "\"-\"", "null", -1)

		df := dataframe.ReadJSON(strings.NewReader(str), dataframe.WithTypes(map[string]series.Type{
			"f12": series.String, "f13": series.String,
		})).RenameDic(name)

		AllStock[marketType] = calData(df, marketType)

		MyChan <- true

		for !common.IsOpen(marketType) {
			time.Sleep(time.Millisecond * 500)
		}
		if marketType == "CNIndex" {
			time.Sleep(time.Second * 1)
		}
		time.Sleep(time.Millisecond * 200)
	}
}

// GoDownload 下载函数
func GoDownload() {
	go getEastMoney("CN")
	go getEastMoney("CNIndex")
	go getEastMoney("HK")
	go getEastMoney("US")
	go getEastMoney("US")
}
