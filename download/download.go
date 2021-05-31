package download

import (
	"context"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strings"
	"test/common"
	"test/marketime"
	"time"
)

const (
	URL = "https://push2.eastmoney.com/api/qt/clist/get?"
)

var ctx = context.Background()
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// MyChan 通道
var MyChan = make(chan bool)

// 计算股票指标
func setStockData(stocks []bson.M, marketType string) []bson.M {
	var results []bson.M
	for _, s := range stocks {
		//代码格式
		if marketType == "CN" {
			if s["code"].(string)[0] == '6' {
				s["code"] = s["code"].(string) + ".SH"
			} else {
				s["code"] = s["code"].(string) + ".SZ"
			}
			// 指数
		} else if marketType == "CNIndex" {
			if s["code"].(string)[0] == '0' {
				s["code"] = s["code"].(string) + ".SH"
			} else {
				s["code"] = s["code"].(string) + ".SZ"
			}
			s["marketType"] = "CN"
			s["type"] = "index"
			s["_id"] = s["code"]
			goto APP
		} else {
			s["code"] = s["code"].(string) + "." + marketType
		}
		s["_id"] = s["code"]
		// 是股票
		if s["total_share"].(float64) > 0 {
			s["type"] = "stock"
			s["marketType"] = marketType
			s["main_net"] = s["main_huge"].(float64) + s["main_big"].(float64)
			s["main_in"] = s["main_net"]
			s["main_out"] = s["main_net"]
			// 总市值，流通市值，换手率在前端实时计算
		}
	APP:
		results = append(results, s)
	}
	return results
}

// 下载数据
func getEastMoney(marketType string) {
	fs := map[string]string{
		"CNIndex": "fs=m:1+s:2,m:0+t:5",                         //沪深指数
		"CN":      "fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",       // 沪深
		"HK":      "fs=m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4", // 港股
		"US":      "fs=m:105,m:106,m:107",                       // 美股
	}
	url := URL + "po=1&fid=f6&pz=6000&np=1&fltt=2&pn=1" + "&" + fs[marketType] + "&fields="
	// 重命名
	var rename bson.M
	if marketType != "CNIndex" {
		rename = bson.M{
			"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
			"f17": "open", "f12": "code", "f10": "vr", "f13": "cid", "f14": "name", "f18": "close",
			"f22": "涨速", "f23": "pb", "f33": "wb",
			// "f34": "外盘", "f35": "内盘", "f24": "pct60day", "f25": "pct_current_year", "f11": "pct5min",
			"f38": "total_share", "f39": "float_share", "f115": "pe_ttm", "f100": "EMIds",
			// 财务
			// "f37": "roe", "f40": "营收", "f41": "营收同比", "f45": "净利润", "f46": "净利润同比",
			// 资金
			"f66": "main_huge", "f72": "main_big", "f78": "main_mid", "f84": "main_small", "f184": "main_pct",
		}
	} else {
		rename = bson.M{
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
			panic(err)
		}
		str := json.Get(body, "data", "diff").ToString()
		//改名
		for i, item := range rename {
			str = strings.Replace(str, i+"\"", item.(string)+"\"", -1)
		}
		// json解析
		var temp []bson.M
		_ = json.Unmarshal([]byte(str), &temp)
		// 计算数据
		temp = setStockData(temp, marketType)
		writeToMongo(temp)
		// 更新完成后传入通道
		MyChan <- true
		for !marketime.IsOpen(marketType) {
			time.Sleep(time.Millisecond * 100)
		}
		time.Sleep(time.Second * 3)
	}
}

// GoDownload 主下载函数
func GoDownload() {
	go getEastMoney("CN")
	go getEastMoney("CNIndex")
	go getEastMoney("HK")
	go getEastMoney("US")
}
