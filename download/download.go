package download

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"test/marketime"
	"time"
)

// redis数据库
var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary

/* 计算股票指标 */
func setStockData(stocks []map[string]interface{}) []map[string]interface{} {
	for i := range stocks {
		s := stocks[i]
		//代码格式
		if s["code"].(string)[0] == '6' {
			s["code"] = s["code"].(string) + ".SH"
		} else {
			s["code"] = s["code"].(string) + ".SZ"
		}
		var 万 float64 = 10000
		labels := []string{"总股本", "流通股本", "特大单流入", "特大单流出", "大单流入", "大单流出", "中单净流入", "小单净流入"}
		for i := range labels {
			col := labels[i]
			s[col] = s[col].(float64) / 万
		}
		// 主力资金
		s["特大单净流入"] = s["特大单流入"].(float64) - s["特大单流出"].(float64)
		s["大单净流入"] = s["大单流入"].(float64) - s["大单流出"].(float64)
		s["主力流入"] = s["大单流入"].(float64) + s["特大单流入"].(float64)
		s["主力流出"] = s["大单流出"].(float64) + s["特大单流出"].(float64)
		s["主力净流入"] = s["主力流入"].(float64) - s["主力流出"].(float64)
		// 市值
		s["总市值"] = s["price"].(float64) * s["总股本"].(float64) / 万
		s["流通市值"] = s["price"].(float64) * s["流通股本"].(float64) / 万
		// 其他
		s["change"] = s["price"].(float64) - s["close"].(float64)
		s["换手率"] = s["vol"].(float64) / s["总股本"].(float64)
		// 保留两位小数
	}
	return stocks
}

/* 下载所有股票数据 */
func getStock(page int) {
	url := "https://push2.eastmoney.com/api/qt/clist/get?pz=2250&np=1&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fltt=2"
	url += "&pn=" + strconv.Itoa(page) + "&fields="
	// 重命名
	nameMaps := map[string]string{
		"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "量比", "f11": "5min涨幅", "f18": "close", "f22": "涨速",
		"f23": "pb", "f24": "60日涨幅", "f25": "年初至今涨跌幅", "f33": "委比",
		"f34": "外盘", "f35": "内盘", "f38": "总股本", "f39": "流通股本", "f115": "pe_ttm",
		// 财务
		// "f37": "roe", "f40": "营收", "f41": "营收同比", "f45": "净利润", "f46": "净利润同比",
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
	for {
		start := time.Now()
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
		// 定义存储类型
		var stocks []map[string]interface{}
		err = json.Unmarshal([]byte(str), &stocks)
		// 计算数据
		stocks = setStockData(stocks)

		// 并发计算排行榜
		var wg = sync.WaitGroup{}
		wg.Add(1)
		go func() {
			ranks(stocks)
			wg.Done()
		}()

		for i := range stocks {
			s := stocks[i]
			// 去掉退市
			if s["总市值"].(float64) == 0 {
				continue
			}
			// 写入数据
			err := rdb.HMSet(ctx, s["code"].(string), s).Err()
			if err != nil {
				log.Println(err)
			}
		}
		// 等待释放
		wg.Wait()

		cost := time.Since(start)
		fmt.Printf("Stocks = %s\n", cost)

		// 当前闭市
		for !marketime.IsOpen() {
			time.Sleep(time.Second * 1)
			continue
		}
	}
}

/* 下载所有指数数据 */
func getIndex() {
	url := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=5000&po=1&np=1&fs=m:1+s:2,m:0+t:5&fltt=2&fields="
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
	for {
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
		// 转maps 直接存取
		var index []map[string]interface{}
		err = json.Unmarshal([]byte(str), &index)

		for i := range index {
			maps := index[i]
			//去掉4 9开头的代码
			if maps["code"].(string)[0] == '4' {
				continue
			} else if maps["code"].(string)[0] == '9' {
				continue
			}
			// 代码格式化
			if maps["code"].(string)[0] == '0' {
				maps["code"] = maps["code"].(string) + ".SH"
			} else {
				maps["code"] = maps["code"].(string) + ".SZ"
			}
			// 写入数据
			err := rdb.HMSet(ctx, maps["code"].(string), maps).Err()
			if err != nil {
				log.Println(err)
			}
		}
		// 间隔3秒 指数不需要高频更新
		time.Sleep(time.Second * 3)
		// 当前闭市
		for !marketime.IsOpen() {
			time.Sleep(time.Second * 1)
			continue
		}
	}
}

func GoDownload() {
	// 主下载函数
	go getStock(1)
	go getStock(2)
	go getIndex()
}
