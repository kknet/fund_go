package download

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"test/marketime"
	"time"
)

var ctx = context.Background()

// redis数据库
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

type Stock struct {
	Price      float32 `json:"price"`
	PctChg     float32 `json:"pct_chg"`
	PreClose   float32 `json:"昨收"`
	Vol        float64 `json:"vol"`
	Amount     float64 `json:"amount"`
	Amp        float32 `json:"amp"`
	Vr         float32 `json:"量比"`
	FiveMin    float32 `json:"5min涨幅"`
	Code       string  `json:"code"`
	High       float32 `json:"high"`
	Low        float32 `json:"low"`
	Open       float32 `json:"open"`
	F1         float32 `json:"涨速"`
	Pb         float32 `json:"pb"`
	PeTtm      float32 `json:"pe_ttm"`
	F24        float32 `json:"60日涨幅"`
	F25        float32 `json:"年初至今涨跌幅"`
	WeiBi      float32 `json:"委比"`
	F34        int     `json:"外盘"`
	F35        int     `json:"内盘"`
	Roe        float32 `json:"roe"`
	TotalShare float64 `json:"总股本"`
	FloatShare float64 `json:"流通股本"`
	F40        float64 `json:"营收"`
	F41        float32 `json:"营收同比"`
	F45        float64 `json:"净利润"`
	F46        float32 `json:"净利润同比"`
	TeIn       float64 `json:"特大单流入"`
	TeOut      float64 `json:"特大单流出"`
	BigIn      float64 `json:"大单流入"`
	BigOut     float64 `json:"大单流出"`
	MidNet     float64 `json:"中单净流入"`
	SmallNet   float64 `json:"小单净流入"`
	// 以下为需要实时计算的数据
	Change       float32 `json:"涨跌"`
	TurnoverRate float32 `json:"换手率"`
	TeNet        float64 `json:"特大单净流入"`
	BigNet       float64 `json:"大单净流入"`
	TotalMkt     float64 `json:"总市值"`
	FloatMkt     float64 `json:"流通市值"`
}

func setStockData(s *Stock) { // 计算
	//代码格式
	if s.Code[0] == '6' {
		s.Code += ".SH"
	} else {
		s.Code += ".SZ"
	}
	s.Change = s.Price - s.PreClose
	s.TurnoverRate = float32(s.Vol / s.TotalShare)
	s.TeNet = s.TeIn - s.TeOut
	s.BigNet = s.BigIn - s.BigOut
	s.TotalMkt = float64(s.Price) * s.TotalShare
	s.FloatMkt = float64(s.Price) * s.FloatShare
}

type EastMoneyData struct {
	Data struct {
		Stocks []Stock `json:"diff"`
	} `json:"data"`
}

func GetStock(page int) {
	url := "https://push2.eastmoney.com/api/qt/clist/get?pz=1500&np=1&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fltt=2"
	url += "&pn=" + strconv.Itoa(page) + "&fields="
	// 重命名
	nameMaps := map[string]string{
		"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "量比", "f11": "5min涨幅", "f18": "昨收", "f22": "涨速",
		"f23": "pb", "f24": "60日涨幅", "f25": "年初至今涨跌幅", "f33": "委比",
		"f34": "外盘", "f35": "内盘", "f38": "总股本", "f39": "流通股本", "f115": "pe_ttm",
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
	for {
		start := time.Now()
		// 从东方财富下载数据
		res, err := http.Get(url)
		if err != nil {
			log.Println(err)
		}
		// 关闭连接
		defer res.Body.Close()
		// 读取内容
		body, err := ioutil.ReadAll(res.Body)
		str := string(body)
		for i := range nameMaps {
			str = strings.Replace(str, i+"\":", nameMaps[i]+"\":", -1)
		}
		// json解析
		info := &EastMoneyData{}
		err = json.Unmarshal([]byte(str), info)

		for i := range info.Data.Stocks {
			s := info.Data.Stocks[i]
			// 计算数据
			setStockData(&s)
			// 退市
			if s.Price == 0 {
				continue
			}
			// 先转换成json 再json转换成map
			data, _ := json.Marshal(&s)
			maps := make(map[string]interface{})
			_ = json.Unmarshal(data, &maps)

			err := rdb.HMSet(ctx, s.Code, maps).Err()
			if err != nil {
				log.Println(err)
			}
		}
		cost := time.Since(start)
		fmt.Printf("EestMoney = %s\n", cost)

		// 当前闭市
		for !marketime.IsOpen() {
			time.Sleep(time.Second * 1)
			continue
		}
	}
}

func GetIndex() {
	url := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=5000&po=1&np=1&fs=m:1+s:2,m:0+t:5&fltt=2&fields="
	// 重命名
	nameMaps := map[string]string{
		"f2": "price", "f3": "pct_chg", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "量比", "f11": "5min涨幅", "f18": "昨收", "f22": "涨速",
		"f23": "pb", "f24": "60日涨幅", "f25": "年初至今涨跌幅", "f33": "委比",
		"f34": "外盘", "f35": "内盘", "f38": "总股本", "f39": "流通股本", "f115": "pe_ttm",
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
	for {
		start := time.Now()
		// 从东方财富下载数据
		res, err := http.Get(url)
		if err != nil {
			log.Println(err)
		}
		// 关闭连接
		defer res.Body.Close()
		// 读取内容
		body, err := ioutil.ReadAll(res.Body)
		str := string(body)
		for i := range nameMaps {
			str = strings.Replace(str, i+"\":", nameMaps[i]+"\":", -1)
		}
		// json解析
		info := &EastMoneyData{}
		err = json.Unmarshal([]byte(str), info)

		for i := range info.Data.Stocks {
			s := info.Data.Stocks[i]
			// 计算数据
			setStockData(&s)
			// 退市
			if s.Price == 0 {
				continue
			}
			// 先转换成json 再json转换成map
			data, _ := json.Marshal(&s)
			maps := make(map[string]interface{})
			_ = json.Unmarshal(data, &maps)

			err := rdb.HMSet(ctx, s.Code, maps).Err()
			if err != nil {
				log.Println(err)
			}
		}
		cost := time.Since(start)
		fmt.Printf("EestMoney = %s\n", cost)

		// 当前闭市
		for !marketime.IsOpen() {
			time.Sleep(time.Second * 1)
			continue
		}
	}
}
