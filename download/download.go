package download

import (
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

//通道
var myChan chan bool

// AllStock 存储所有股票数据
var AllStock = map[string]dataframe.DataFrame{}

// 计算指标
func calData(df dataframe.DataFrame, marketType string) dataframe.DataFrame {
	length := df.Nrow()

	//删除所有值为 "0" 的列
	for _, col := range df.Names() {
		s := df.Col(col)
		if s.Max() == 0 {
			df = df.Drop(s.Name)
		}
	}
	//代码格式化
	code := df.Select("code").Capply(func(s series.Series) series.Series {
		for i := 0; i < s.Len(); i++ {
			c := s.Elem(i)
			str := c.String()
			switch marketType {
			case "CN":
				c.Set(str + Expression(str[0] == '6', ".SH", ".SZ").(string))
			case "CNIndex":
				c.Set(str + Expression(str[0] == '0', ".SH", ".SZ").(string))
			case "HK", "US":
				c.Set(str + "." + marketType)
			}
		}
		return s
	}).Col("code")
	df = df.Mutate(code)

	df = df.Mutate(newSeries(Expression(marketType == "CNIndex", "CN", marketType), "marketType", length))
	df = df.Mutate(newSeries(Expression(marketType == "CNIndex", "index", "stock"), "type", length))

	// 计算涨跌幅
	pct := df.Select([]string{"price", "close"}).Rapply(func(s series.Series) series.Series {
		pctChg := (s.Elem(0).Float()/s.Elem(1).Float() - 1.0) * 100
		s.Elem(0).Set(pctChg)
		return s
	}).Rename("pct_chg", "X0").Col("pct_chg")
	df = df.Mutate(pct)

	// 去除涨跌幅为空的数据
	for i := 0; i < pct.Len(); i++ {
		if pct.Elem(i).IsNA() {
			df = df.Drop(i)
		}
	}
	// 计算换手率 市值 资金净流入
	if marketType != "CNIndex" {
		data := df.Select([]string{"total_share", "float_share", "price", "vol", "内盘", "外盘"}).Rapply(func(s series.Series) series.Series {
			mc := s.Elem(0).Float() * s.Elem(2).Float()
			fmc := s.Elem(1).Float() * s.Elem(2).Float()
			tr := s.Elem(3).Float() / s.Elem(0).Float() * 10000
			net := (s.Elem(5).Float() - s.Elem(4).Float()) * s.Elem(2).Float()

			s.Set([]int{0, 1, 2, 3}, series.Floats([]float64{mc, fmc, tr, net}))
			return s
		}).Rename("mc", "X0").Rename("fmc", "X1").
			Rename("tr", "X2").Rename("net", "X3").Drop([]string{"X4", "X5"})
		df = df.CBind(data)
	}
	//主力资金流向
	if marketType == "CN" {
		data := df.Select([]string{"main_huge", "main_big", "main_pct"}).Rapply(func(s series.Series) series.Series {
			net := s.Elem(0).Float() + s.Elem(1).Float()
			amount := net / s.Elem(2).Float() * 100
			in := (net + amount) / 2.0
			out := net - in

			s.Set([]int{0, 1, 2}, series.Floats([]float64{net, in, out}))
			return s
		}).Rename("main_net", "X0").Rename("main_in", "X1").Rename("main_out", "X2")
		df = df.CBind(data)
	}
	return df
}

// 下载数据
func getEastMoney(marketType string) {
	fs := map[string]string{
		"CNIndex": "m:1+s:2,m:0+t:5",
		"CN":      "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23",
		"HK":      "m:116+t:1,m:116+t:2,m:116+t:3,m:116+t:4",
		"US":      "m:105,m:106,m:107",
	}
	url := URL + "po=1&fid=f6&pz=8000&np=1&fltt=2&pn=1&fs=" + fs[marketType] + "&fields="
	// 重命名map
	rename := map[string]string{
		"f2": "price", "f5": "vol", "f6": "amount", "f7": "amp", "f15": "high", "f16": "low",
		"f17": "open", "f12": "code", "f10": "vr", "f13": "cid", "f14": "name", "f18": "close",
		"f23": "pb", "f34": "外盘", "f35": "内盘",
		//"f22": "涨速", "f11": "pct5min", "f24": "pct60day", "f25": "pct_current_year",
		"f38": "total_share", "f39": "float_share", "f115": "pe_ttm",
		//"f100": "EMIds",
		//"f37": "roe", "f40": "营收", "f41": "营收同比", "f45": "净利润", "f46": "净利润同比",
	}
	if marketType == "CN" {
		rename["f33"] = "wb"
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
		body, err := request.Do()
		if err != nil {
			log.Println("下载股票数据发生错误，", err.Error())
		}
		str := json.Get(body, "data", "diff").ToString()
		//初始化
		df := dataframe.ReadJSON(strings.NewReader(str), dataframe.WithTypes(map[string]series.Type{
			"f12": series.String, "f13": series.String,
		}))
		//改名
		for key, value := range rename {
			df = df.Rename(value, key)
		}
		//计算
		AllStock[marketType] = calData(df, marketType)

		for !common.IsOpen(marketType) {
			time.Sleep(time.Millisecond * 500)
		}
		time.Sleep(time.Millisecond * 300)
	}
}

// GoDownload 主下载函数
func GoDownload() {
	go getEastMoney("CN")
	go getEastMoney("CNIndex")
	go getEastMoney("HK")
	go getEastMoney("US")
	go getEastMoney("US")
}
