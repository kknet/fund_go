package download

import (
	"context"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strings"
	"test/marketime"
	"time"
)

// redis数据库
var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0,
})

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var AllStock []map[string]interface{}
var AllIndex []map[string]interface{}

/* 计算股票指标 */
func setStockData(stocks []map[string]interface{}) {
	for i := range stocks {
		s := stocks[i]
		//代码格式
		if s["code"].(string)[0] == '6' {
			s["code"] = s["code"].(string) + ".SH"
		} else {
			s["code"] = s["code"].(string) + ".SZ"
		}
		// 单位 万
		labels := []string{"总股本", "流通股本", "特大单流入", "特大单流出", "大单流入", "大单流出", "中单净流入", "小单净流入"}
		for i := range labels {
			col := labels[i]
			s[col] = s[col].(float64) / 10000
		}
		// 主力资金
		s["特大单净流入"] = s["特大单流入"].(float64) - s["特大单流出"].(float64)
		s["大单净流入"] = s["大单流入"].(float64) - s["大单流出"].(float64)
		s["主力流入"] = s["大单流入"].(float64) + s["特大单流入"].(float64)
		s["主力流出"] = s["大单流出"].(float64) + s["特大单流出"].(float64)
		s["主力净流入"] = s["主力流入"].(float64) - s["主力流出"].(float64)
		// 市值
		s["总市值"] = s["price"].(float64) * s["总股本"].(float64) / 10000
		s["流通市值"] = s["price"].(float64) * s["流通股本"].(float64) / 10000
		// 其他
		s["change"] = s["price"].(float64) - s["close"].(float64)
		s["换手率"] = s["vol"].(float64) / s["总股本"].(float64)

		s["资金净流入"] = s["外盘"].(float64) - s["内盘"].(float64)
		// 保留两位小数
	}
}

/* 下载所有股票数据 */
func getStock() {
	url := "https://push2.eastmoney.com/api/qt/clist/get?pz=5000&np=1&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23"
	// 按成交额降序
	url += "&fltt=2&po=1&fid=f6&pn=1&fields="
	// 重命名
	nameMaps := map[string]string{
		"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "量比", "f11": "5min涨幅", "f14": "name", "f18": "close",
		"f22": "涨速", "f23": "pb", "f33": "委比",
		"f24": "60日涨幅", "f25": "年初至今涨跌幅", "f34": "外盘", "f35": "内盘",
		"f38": "总股本", "f39": "流通股本", "f115": "pe_ttm",
		// 财务
		"f37": "roe", "f40": "营收", "f41": "营收同比", "f45": "净利润", "f46": "净利润同比",
		// 资金
		"f64": "特大单流入", "f65": "特大单流出",
		"f70": "大单流入", "f71": "大单流出",
		"f78": "中单净流入", "f84": "小单净流入",
	}
	//连接参数
	for i := range nameMaps {
		url += i + ","
	}
	//去掉末尾的逗号
	url = url[:len(url)-1]
	// 从东方财富下载数据
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	str := json.Get(body, "data", "diff").ToString()
	//改名
	for i := range nameMaps {
		str = strings.Replace(str, i+"\"", nameMaps[i]+"\"", -1)
	}
	// json解析
	var temp []map[string]interface{}
	err = json.Unmarshal([]byte(str), &temp)
	// 计算数据
	setStockData(temp)
	AllStock = temp
}

/* 下载所有指数数据 */
func getIndex() {
	url := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=5000&po=1&np=1&fs=m:1+s:2,m:0+t:5"
	url += "&fltt=2&fields="
	// 重命名
	nameMaps := map[string]string{
		"f2": "price", "f3": "pct_chg", "f4": "change", "f5": "vol", "f6": "amount", "f7": "amp", "f8": "换手率",
		"f14": "name", "f15": "high", "f16": "low", "f17": "open", "f18": "close", "f12": "code", "f11": "5min涨幅",
	}
	//连接参数
	for i := range nameMaps {
		url += i + ","
	}
	//去掉末尾的逗号
	url = url[:len(url)-1]
	// 从东方财富下载数据
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	str := json.Get(body, "data", "diff").ToString()
	//改名
	for i := range nameMaps {
		str = strings.Replace(str, i+"\"", nameMaps[i]+"\"", -1)
	}
	// json解析
	var temp []map[string]interface{}
	err = json.Unmarshal([]byte(str), &temp)
	for i := range temp {
		s := temp[i]
		//代码格式
		if s["code"].(string)[0] == '0' {
			s["code"] = s["code"].(string) + ".SH"
		} else {
			s["code"] = s["code"].(string) + ".SZ"
		}
	}
	AllIndex = temp
}

// MyChannel 定义通道
var MyChannel = make(chan bool)

// GoDownload 主下载函数
func GoDownload() {
	// 股票
	go func() {
		for {
			getStock()
			for !marketime.IsOpen() {
			}
			// 更新完成后传入通道
			MyChannel <- true
			// 1秒
			time.Sleep(time.Second * 1)
		}
	}()
	// 指数
	go func() {
		for {
			getIndex()
			for !marketime.IsOpen() {
			}
			time.Sleep(time.Second * 3)
		}
	}()
	// 求实时排行榜
	go ranks(AllStock)
	// 求市场数据
	//go getMarketData()
}