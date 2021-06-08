package download

import (
	"fund_go2/common"
	"github.com/go-gota/gota/dataframe"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
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

// 计算股票指标
func setStockData(stocks []bson.M, marketType string) {
	for _, s := range stocks {
		//代码格式
		switch marketType {

		case "CN":
			if s["code"].(string)[0] == '6' {
				s["code"] = s["code"].(string) + ".SH"
			} else {
				s["code"] = s["code"].(string) + ".SZ"
			}
			s["marketType"] = "CN"
			s["type"] = "stock"
			s["_id"] = s["code"]

		case "CNIndex":
			if s["code"].(string)[0] == '0' {
				s["code"] = s["code"].(string) + ".SH"
			} else {
				s["code"] = s["code"].(string) + ".SZ"
			}
			s["marketType"] = "CN"
			s["type"] = "index"
			s["_id"] = s["code"]
			continue

		case "HK", "US":
			s["code"] = s["code"].(string) + "." + marketType
			s["marketType"] = marketType
			s["type"] = "stock"
			s["_id"] = s["code"]
		}

		// 是股票
		if s["total_share"].(float64) > 0 {
			s["main_net"] = s["main_huge"].(float64) + s["main_big"].(float64)
			s["main_in"] = s["main_net"]
			s["main_out"] = s["main_net"]
			s["mc"] = s["total_share"].(float64) * s["price"].(float64)
			s["fmc"] = s["float_share"].(float64) * s["price"].(float64)
		}
	}
}

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
		"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "vr", "f13": "cid", "f14": "name", "f18": "close",
		"f23": "pb", "f33": "wb",
		"f34": "外盘", "f35": "内盘", "f22": "涨速", "f11": "pct5min", "f24": "pct60day", "f25": "pct_current_year",
		"f38": "total_share", "f39": "float_share", "f115": "pe_ttm",
		"f100": "EMIds",
		// 财务
		//"f37": "roe", "f40": "营收", "f41": "营收同比", "f45": "净利润", "f46": "净利润同比",
		// 资金
		"f66": "main_huge", "f72": "main_big", "f78": "main_mid", "f84": "main_small", "f184": "main_pct",
	}
	if marketType == "CNIndex" {
		rename = map[string]string{
			"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
			"f17": "open", "f12": "code", "f14": "name", "f18": "close", "f8": "tr", "f13": "cid",
		}
	}
	//连接参数
	for i := range rename {
		url += i + ","
	}
	//去掉末尾的逗号
	url = url[:len(url)-1]

	request := common.NewGetRequest(url)
	for {
		body, err := request.Do()
		if err != nil {
			log.Println("下载股票数据发生错误，", err.Error())
		}
		str := json.Get(body, "data", "diff").ToString()

		CNStock = dataframe.ReadJSON(strings.NewReader(str))
		//改名
		for key, value := range rename {
			CNStock = CNStock.Rename(value, key)
		}

		for !common.IsOpen(marketType) {
			time.Sleep(time.Millisecond * 100)
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

// GoDownload 主下载函数
func GoDownload() {
	go getEastMoney("CN")
	//go getEastMoney("CNIndex")
	//go getEastMoney("HK")
	//go getEastMoney("US")
}
