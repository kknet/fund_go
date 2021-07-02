package download

import (
	"fmt"
	"fund_go2/common"
	jsoniter "github.com/json-iterator/go"
	"log"
	"strconv"
	"sync"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var Industry = map[string]interface{}{}

// MyChan 全局通道
var MyChan = getGlobalChan()

// 将MyChan设置为全局通道
func getGlobalChan() chan string {
	var ch chan string
	var chanOnceManager sync.Once

	chanOnceManager.Do(func() {
		ch = make(chan string)
	})
	return ch
}

// 计算指标
func calData(stocks []Stock, marketType string) []Stock {
	for i := range stocks {
		s := &stocks[i]
		// 基本信息
		s.Cid = strconv.Itoa(s.C) + "." + s.Code
		switch marketType {
		case "CN":
			s.Code += Expression(s.Code[0] == '6', ".SH", ".SZ").(string)
		case "Index":
			s.Code += Expression(s.Code[0] == '0', ".SH", ".SZ").(string)
		case "HK", "US":
			s.Code += "." + marketType
		}
		s.Id = s.Code
		s.MarketType = Expression(marketType == "Index", "CN", marketType).(string)
		s.Type = Expression(marketType == "Index", "index", "stock").(string)

		// 计算指标
		if s.Close > 0 {
			s.PctChg = (s.Price/s.Close - 1) * 100
			s.Amp = (s.High - s.Low) / s.Close * 100
		} else {
			continue
		}
		// mc fmc tr
		if s.TotalShare > 0 {
			s.Mc = s.TotalShare * s.Price
			s.Fmc = s.FloatShare * s.Price
			s.Tr = s.Vol / s.TotalShare * 10000
		}
		// net
		if s.Buy > 0 && s.Sell > 0 {
			s.Net = (s.Buy - s.Sell) * s.Amount / s.Vol
		}
		// 主力净流入
		if marketType == "CN" && s.MainPct != 0 {
			s.MainNet = s.MainHuge + s.MainBig
			amount := s.MainNet / s.MainPct * 100
			s.MainIn = (s.MainNet + amount) / 2
			s.MainOut = s.MainNet - s.MainIn
		}
	}
	return stocks
}

// 下载数据
func getEastMoney(marketType string, page int) {
	fs := map[string]string{
		"Index": "m:1+s:2,m:0+t:5",
		"CN":    "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",
		"HK":    "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4",
		"US":    "m:105,m:106,m:107",
	}
	url := "https://push2.eastmoney.com/api/qt/clist/get?po=1&fid=f6&pz=2500&np=1&fltt=2&pn=" +
		strconv.Itoa(page) + "&fs=" + fs[marketType] + "&fields="
	// 重命名
	name := map[string]string{
		"f2": "price", "f5": "vol", "f6": "amount", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "vr", "f13": "cid", "f14": "name", "f18": "close",
		"f23": "pb", "f34": "外盘", "f35": "内盘",
		"f22": "pct_rate", "f11": "pct5min", "f24": "pct60day", "f25": "pct_year",
		"f38": "total_share", "f39": "float_share", "f115": "pe_ttm",
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

			fmt.Println("正在重试...")
			request = common.NewGetRequest(url)
			continue
		}
		var info AutoGenerated
		_ = json.Unmarshal(body, &info)

		data := calData(info.Data.Diff, marketType)
		UpdateMongo(data, marketType)

		if marketType == "CN" {
			go CalIndustry()
		}
		MyChan <- marketType

		for !common.IsOpen(marketType) {
			time.Sleep(time.Millisecond * 500)
		}
		time.Sleep(time.Millisecond * 300)
	}
}

// GoDownload 下载函数
func GoDownload() {
	UpdateBasic()
	go getEastMoney("CN", 1)
	go getEastMoney("CN", 2)
	go getEastMoney("Index", 1)
	go getEastMoney("HK", 1)
	go getEastMoney("US", 1)
	go getEastMoney("US", 2)
	go getEastMoney("US", 3)
}
