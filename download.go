package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var ctx = context.Background()

// redis数据库
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func isOpen() bool { //是否开市
	// 工作日
	if time.Now().Weekday() < 5 {
		// 上午
		if time.Now().Hour() < 9 || time.Now().Hour() >= 15 {
			return false
		}
	}
	return false
}

type XueQiuData struct {
	Data struct {
		Count int `json:"count"` // 总数
		List  []struct {
			Symbol             string  `json:"symbol"`          //代码
			NetProfitCagr      float32 `json:"net_profit_cagr"` //净利润率
			Percent            float32 `json:"percent"`         //涨跌幅
			PbTtm              float32 `json:"pb_ttm"`          //pb_ttm
			Current            float32 `json:"current"`         //价格
			Amplitude          float32 `json:"amplitude"`       //振幅
			Pcf                float32 `json:"pcf"`
			CurrentYearPercent float32 `json:"current_year_percent"` //近一年涨幅
			DividendYield      int     `json:"dividend_yield"`       //股息收益率 股息率%
			RoeTtm             float32 `json:"roe_ttm"`              //roe_ttm
			Percent5M          float32 `json:"percent5m"`            //5min涨幅
			IncomeCagr         float32 `json:"income_cagr"`
			Amount             float32 `json:"amount"`               //成交额
			MainNetInflows     int     `json:"main_net_inflows"`     //资金净流入
			Volume             int     `json:"volume"`               //成交量
			VolumeRatio        int     `json:"volume_ratio"`         //量比
			Pb                 float32 `json:"pb"`                   //pb
			Followers          int     `json:"followers"`            //总关注
			TurnoverRate       float32 `json:"turnover_rate"`        //换手率
			Name               string  `json:"name"`                 //名字
			PeTtm              float32 `json:"pe_ttm"`               //pe_ttm
			MarketCapital      int     `json:"market_capital"`       //总市值
			FloatMarketCapital int     `json:"float_market_capital"` //流通市值
			TotalShares        int     `json:"total_shares"`         //总股本
			FloatShares        int     `json:"float_shares"`         //流通股本
		} `json:"list"`
	} `json:"data"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

type EastMoneyData struct {
	Data struct {
		Total int `json:"total"`
		Diff  []struct {
			Code     string  `json:"code"`
			High     float32 `json:"high"`
			Low      float32 `json:"low"`
			Open     float32 `json:"open"`
			PreClose float32 `json:"pre_close"`
			WeiBi    float32 `json:"委比"`    //委比
			Buy      float32 `json:"外盘"`    //外盘（买）
			Sell     float32 `json:"内盘"`    //内盘（卖）
			NetMain  float32 `json:"主力净流入"` // 主力净流入
			BuyHuge  float32 `json:"特大单流入"`
			SellHuge float32 `json:"特大单流出"`
			NetHuge  float32 `json:"特大单净流入"`
			BuyBig   float32 `json:"大单流入"`
			SellBig  float32 `json:"大单流出"`
			NetBig   float32 `json:"大单净流入"`
			NetMid   float32 `json:"净中单"` //中单
			NetSmall float32 `json:"净小单"` // 小单净流入
		} `json:"diff"`
	} `json:"data"`
}

func DownloadXueQiu() {
	url := "https://xueqiu.com/service/v5/stock/screener/quote/list?size=5000&order_by=percent&type=sh_sz"
	// 从雪球下载数据
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, _ := ioutil.ReadAll(res.Body)
	// json解析
	info := &XueQiuData{}
	err = json.Unmarshal(body, info)

	start := time.Now()
	for i := range info.Data.List {
		temp := info.Data.List[i]
		temp.Symbol = temp.Symbol[2:] + "." + temp.Symbol[0:2]
		//去掉B股
		if temp.Symbol[0] == '9' {
			continue
		}
		// 先转换成json 再json转换成map
		data, _ := json.Marshal(&temp)
		maps := make(map[string]interface{})
		_ = json.Unmarshal(data, &maps)

		err := rdb.HMSet(ctx, temp.Symbol, maps).Err()
		if err != nil {
			fmt.Println(err)
		}
	}
	cost := time.Since(start)
	fmt.Printf("redis = %s\n", cost)
}

func DownloadEastMoney() {
	url := "https://push2.eastmoney.com/api/qt/clist/get?pn=1&pz=5000&np=1&fs=m:0+t:6,m:0+t:13,m:0+t:80,m:1+t:2,m:1+t:23&fltt=2&fields="
	// 重命名
	nameMaps := map[string]string{
		"f12": "code", "f15": "high", "f16": "low", "f17": "open", "f18": "pre_close", "f33": "委比", "f34": "外盘", "f35": "内盘",
		"f62": "主力净流入",
		"f64": "特大单流入", "f65": "特大单流出", "f66": "特大单净流入",
		"f70": "大单流入", "f71": "大单流出", "f72": "大单净流入",
		"f78": "净中单", "f84": "净小单",
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
		log.Println(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	str := string(body)
	for i := range nameMaps {
		str = strings.Replace(str, i, nameMaps[i], -1)
	}
	// json解析
	info := &EastMoneyData{}
	err = json.Unmarshal([]byte(str), info)

	for i := range info.Data.Diff {
		temp := info.Data.Diff[i]
		if temp.Code[0] == '6' {
			temp.Code += ".SH"
		} else {
			temp.Code += ".SZ"
		}
		// 退市
		if temp.High == 0 {
			continue
		}
		// 先转换成json 再json转换成map
		data, _ := json.Marshal(&temp)
		maps := make(map[string]interface{})
		_ = json.Unmarshal(data, &maps)

		err := rdb.HMSet(ctx, temp.Code, maps).Err()
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	//异步并发
	var wg sync.WaitGroup
	limit := 2

	wg.Add(limit)
	go func() {
		for {
			DownloadXueQiu()
		}
	}()
	go func() {
		for {
			DownloadEastMoney()
		}
	}()
	//并发释放
	wg.Wait()
	fmt.Println("main exit")
}
