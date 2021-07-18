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

// 初始化全局通道
func getGlobalChan() chan string {
	var ch chan string
	var chanOnceManager sync.Once

	chanOnceManager.Do(func() {
		ch = make(chan string)
	})
	return ch
}

// 计算指标
func calData(df dataframe.DataFrame, marketType string) dataframe.DataFrame {

	df = df.Filter(dataframe.F{Colname: "vol", Comparator: series.Greater, Comparando: 0})

	indexes := make([]int, df.Nrow())
	for i := range indexes {
		indexes[i] = i + 1
	}
	index := series.Ints(indexes)

	for _, col := range df.Names() {

		// cid type marketType
		if col == "cid" {
			basic := df.Select([]string{"cid", "code"}).Rapply(func(s series.Series) series.Series {

				cid := s.Elem(0).String() + "." + s.Elem(1).String()
				Type := Expression(marketType == "Index", "index", "stock").(string)

				return series.Strings([]string{cid, marketType, Type})
			})
			_ = basic.SetNames("cid", "marketType", "type")

			for _, c := range basic.Names() {
				df = df.Mutate(basic.Col(c))
			}

			// tr mc fmc
		} else if col == "total_share" {
			pct := df.Select([]string{"price", "vol", "total_share", "float_share"}).Rapply(func(s series.Series) series.Series {
				value := s.Float()

				var tr, mc, fmc float64
				if value[3] > 0 {
					tr = value[1] / value[3] * 10000
				}
				mc = value[0] * value[2]
				fmc = value[0] * value[3]

				return series.Floats([]float64{tr, mc, fmc})
			})
			_ = pct.SetNames("tr", "mc", "fmc")
			df = df.CBind(pct)

			// 主力净流入排名
		} else if strings.Contains(col, "main_net") && (marketType == "CN" || marketType == "HK") {
			df = df.Arrange(dataframe.RevSort(col))
			index.Name = col + "_rank"
			df = df.Mutate(index)
		}
	}

	// code
	code := df.Col("code").Map(func(element series.Element) series.Element {
		code := element.String()
		switch marketType {
		case "CN":
			code += Expression(code[0] == '6', ".SH", ".SZ").(string)
		case "Index":
			code += Expression(code[0] == '0', ".SH", ".SZ").(string)
		case "HK", "US":
			code += "." + marketType
		}
		element.Set(code)
		return element
	})
	df = df.Mutate(code)

	// net
	net := df.Select([]string{"vol", "amount", "buy", "sell"}).Rapply(func(s series.Series) series.Series {
		value := s.Float()
		net := (value[2] - value[3]) * value[1] / value[0]

		return series.Floats([]float64{net})
	}).Col("X0")
	net.Name = "net"
	df = df.Mutate(net).Drop([]string{"buy", "sell"})

	return df
}

// 参数
var fs = map[string]string{
	"Index": "m:1+s:2,m:0+t:5",
	"CN":    "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",
	"HK":    "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4",
	"US":    "m:105,m:106,m:107",
}

// 不需要高频更新
var basicName = map[string]string{
	"f13": "cid", "f14": "name", "f15": "high", "f16": "low", "f17": "open", "f18": "close",
	"f38": "total_share", "f39": "float_share",
	"f267": "3day_main_net", "f164": "5day_main_net", "f174": "10day_main_net",
	"f23": "pb", "f115": "pe_ttm", "f24": "pct60day", "f25": "pct_year",
}

// 需要高频更新
var proName = map[string]string{
	"f12": "code", "f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount",
	"f10": "vr", "f33": "wb", "f34": "buy", "f35": "sell",
	"f22": "pct_rate", "f11": "pct5min",
	"f62": "main_net", "f66": "main_huge", "f72": "main_big", "f78": "main_mid", "f84": "main_small",
}

// 下载数据
func getEastMoney(marketType string) {
	url := "https://push2.eastmoney.com/api/qt/clist/get?po=1&fid=f20&pz=6000&np=1&fltt=2&pn=1&fs=" + fs[marketType] + "&fields="
	var tempUrl string
	// 定时更新计数器
	count := 999
	client := &http.Client{}
	for {
		tempUrl = url
		// 连接参数
		for i := range proName {
			tempUrl += i + ","
		}
		if count >= 10 {
			for i := range basicName {
				tempUrl += i + ","
			}
			count = 0
		}
		tempUrl = tempUrl[:len(tempUrl)-1]

		res, err := client.Get(tempUrl)
		if err != nil {
			log.Println("下载股票数据发生错误，3秒后重试...", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		body, _ := ioutil.ReadAll(res.Body)
		str := json.Get(body, "data", "diff").ToString()
		_ = res.Body.Close()

		df := dataframe.ReadJSON(strings.NewReader(str), dataframe.WithTypes(map[string]series.Type{
			"f12": series.String, "f13": series.String,
		}))

		for _, col := range df.Names() {
			newName, ok := proName[col]
			if !ok {
				newName = basicName[col]
			}
			df = df.Rename(newName, col)
		}

		df = calData(df, marketType)
		UpdateMongo(df.Maps(), marketType)

		if marketType == "CN" {
			// 间隔更新行业数据
			if count%5 == 0 {
				go CalIndustry()
			}
		}
		count++
		MyChan <- marketType

		for !common.IsOpen(marketType) {
			count = 10
			time.Sleep(time.Millisecond * 500)
		}
		time.Sleep(time.Millisecond * 300)
	}
}

// GoDownload 下载函数
func GoDownload() {
	go getEastMoney("CN")
	go getEastMoney("Index")
	go getEastMoney("HK")
	go getEastMoney("US")
}
