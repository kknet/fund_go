package download

import (
	"fmt"
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

	// 价格大于0
	df = df.Filter(dataframe.F{Colname: "price", Comparator: series.Greater, Comparando: 0})

	//基本信息
	basicCol := df.Col("code")
	if &basicCol != nil {
		basic := df.Select([]string{"cid", "code"}).Rapply(func(s series.Series) series.Series {
			var cid, code, Type string

			cid = s.Elem(0).String() + "." + s.Elem(1).String()
			code = s.Elem(1).String()
			switch marketType {
			case "CN":
				Type = "stock"
				code += Expression(code[0] == '6', ".SH", ".SZ").(string)
			case "Index":
				Type = "Index"
				code += Expression(code[0] == '0', ".SH", ".SZ").(string)
			case "HK", "US":
				Type = "stock"
				code += "." + marketType
			}
			return series.Strings([]string{cid, code, marketType, Type})
		})
		_ = basic.SetNames("cid", "code", "marketType", "type")

		for _, col := range basic.Names() {
			df = df.Mutate(basic.Col(col))
		}
	}

	realCol := df.Col("total_share")
	// tr mc fmc
	if &realCol != nil {
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

		for _, col := range pct.Names() {
			df = df.Mutate(pct.Col(col))
		}
	}

	// net
	net := df.Select([]string{"vol", "amount", "buy", "sell"}).Rapply(func(s series.Series) series.Series {
		value := s.Float()
		return series.Floats([]float64{
			(value[2] - value[3]) * value[1] / value[0],
		})
	}).Col("X0")
	net.Name = "net"
	df = df.Mutate(net)

	// main_net 排行

	df = df.Drop([]string{"buy", "sell"})
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
	"f12": "code", "f13": "cid", "f14": "name", "f17": "open", "f18": "close",
	"f38": "total_share", "f39": "float_share",
	"f267": "3day_main_net", "f164": "5day_main_net", "f174": "10day_main_net",
	"f23": "pb", "f115": "pe_ttm",
}

// 需要高频更新
var proName = map[string]string{
	"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f15": "high", "f16": "low",
	"f10": "vr", "f33": "wb", "f34": "buy", "f35": "sell",
	"f22": "pct_rate", "f11": "pct5min", "f24": "pct60day", "f25": "pct_year",
	"f62": "main_net", "f66": "main_huge", "f72": "main_big", "f78": "main_mid", "f84": "main_small",
}

// 下载数据
func getEastMoney(marketType string, page int) {
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?po=1&fid=f6&pz=2500&np=1&fltt=2&pn=%d&fs=%s&fields=", page, fs[marketType])
	// 定时更新计数器
	count := 10
	client := &http.Client{}
	for {
		// 连接参数
		for i := range proName {
			url += i + ","
		}
		if count >= 10 {
			for i := range basicName {
				url += i + ","
			}
			count = 0
		}
		url = url[:len(url)-1]

		res, err := client.Get(url)
		if err != nil {
			log.Println("下载股票数据发生错误，", err.Error())
			log.Println("3秒后重试...")
			time.Sleep(time.Second * 3)
			continue
		}
		body, _ := ioutil.ReadAll(res.Body)
		str := json.Get(body, "data", "diff").ToString()
		_ = res.Body.Close()

		df := dataframe.ReadJSON(strings.NewReader(str), dataframe.WithTypes(map[string]series.Type{
			"f12": series.String, "f13": series.String,
		}))
		for key, value := range basicName {
			df = df.Rename(value, key)
		}
		for key, value := range proName {
			df = df.Rename(value, key)
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
	go getEastMoney("CN", 1)
	go getEastMoney("CN", 2)
	go getEastMoney("Index", 1)
	go getEastMoney("HK", 1)
	go getEastMoney("US", 1)
	go getEastMoney("US", 2)
	go getEastMoney("US", 3)
}
