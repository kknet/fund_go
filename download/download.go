package download

import (
	"fmt"
	"fund_go2/common"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	jsoniter "github.com/json-iterator/go"
	"log"
	"strings"
	"time"
)

const (
	URL = "https://push2.eastmoney.com/api/qt/clist/get?"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// MyChan 通道
var MyChan = make(chan bool)

// CNStock 存储所有股票数据
var CNStock dataframe.DataFrame
var HKStock dataframe.DataFrame
var USStock dataframe.DataFrame

// 下载数据
func getEastMoney(marketType string) {
	fs := map[string]string{
		"CNIndex": "m:1+s:2,m:0+t:5",                         //沪深指数
		"CN":      "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",       // 沪深
		"HK":      "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4", // 港股
		"US":      "m:105,m:106,m:107",                       // 美股
	}
	url := URL + "po=1&fid=f6&pz=8000&np=1&fltt=2&pn=1&fs=" + fs[marketType] + "&fields="
	// 重命名
	rename := map[string]string{
		"f2": "price",
		//"f3": "pct_chg",
		"f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "vr", "f13": "cid", "f14": "name", "f18": "close",
		"f23": "pb", "f33": "wb",
		//"f34": "外盘", "f35": "内盘", "f22": "涨速", "f11": "pct5min", "f24": "pct60day", "f25": "pct_current_year",
		"f38": "total_share", "f39": "float_share", "f115": "pe_ttm", "f100": "EMIds",
		// 财务
		//"f37": "roe", "f40": "营收", "f41": "营收同比", "f45": "净利润", "f46": "净利润同比",
	}
	// 资金
	if marketType == "CN" {
		rename["f66"] = "main_huge"
		rename["f72"] = "main_big"
		rename["f78"] = "main_mid"
		rename["f84"] = "main_small"
		rename["f184"] = "main_pct"
	}
	//连接参数
	for i := range rename {
		url += i + ","
	}
	//去掉末尾的逗号
	url = url[:len(url)-1]

	request := common.NewGetRequest(url)
	for {
		start := time.Now()
		body, err := request.Do()
		if err != nil {
			log.Println("下载股票数据发生错误，", err.Error())
		}
		str := json.Get(body, "data", "diff").ToString()
		df := dataframe.ReadJSON(strings.NewReader(str), dataframe.WithTypes(map[string]series.Type{
			"f12": series.String, "f13": series.String,
		}))
		//改名
		for key, value := range rename {
			df = df.Rename(value, key)
		}
		length := df.Nrow()
		//代码格式化
		df = df.Mutate(newSeries(marketType, "marketType", length))
		df = formatCode(df)

		//删除所有值为 "0" 的列
		for _, col := range df.Names() {
			s := df.Col(col)
			if s.Max() == 0 {
				df = df.Drop(s.Name)
			}
		}
		//计算涨跌幅
		pct := Operation(df.Col("price"), "/", df.Col("close"))
		pct = Operation(pct, "-", 1.0)
		pct = Operation(pct, "*", 100.0)
		pct.Name = "pct_chg"
		df = df.Mutate(pct)

		//计算市值
		mc := Operation(df.Col("total_share"), "*", df.Col("price"))
		mc.Name = "mc"
		df = df.Mutate(mc)

		fmc := Operation(df.Col("float_share"), "*", df.Col("price"))
		fmc.Name = "fmc"
		df = df.Mutate(fmc)
		//计算换手率
		tr := Operation(df.Col("vol"), "/", df.Col("total_share"))
		tr = Operation(tr, "*", 10000.0)
		tr.Name = "tr"
		df = df.Mutate(tr)

		mainNet := Operation(df.Col("main_huge"), "+", df.Col("main_big"))
		mainNet.Name = "main_net"
		df = df.Mutate(mainNet)

		//计算主力流入 主力流出
		mainAmount := Operation(df.Col("main_pct"), "/", 100.0)
		mainAmount = Operation(mainNet, "/", mainAmount)

		t := Operation(mainNet, "+", mainAmount)
		mainIn := Operation(t, "/", 2.0)
		mainIn.Name = "main_in"
		df = df.Mutate(mainIn)

		mainOut := Operation(mainNet, "-", mainIn)
		mainOut.Name = "main_out"
		df = df.Mutate(mainOut)

		fmt.Println(df.Select([]string{"main_in", "main_out", "main_net", "main_pct"}))

		CNStock = df
		fmt.Println("总用时：", time.Since(start))

		for !common.IsOpen(marketType) {
			time.Sleep(time.Millisecond * 100)
		}
		time.Sleep(time.Second * 9999)
	}
}

// GoDownload 主下载函数
func GoDownload() {
	go getEastMoney("CN")
	//go getEastMoney("CNIndex")
	//go getEastMoney("HK")
	//go getEastMoney("US")
}
