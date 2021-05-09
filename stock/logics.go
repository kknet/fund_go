package stock

import (
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strings"
	"test/download"
)

const (
	URL = "https://push2ex.eastmoney.com/getStockFenShi?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wzfscj&pageindex=0&sort=1&ft=1"
)

// jsoniter
var json = jsoniter.ConfigCompatibleWithStandardLibrary

type SourceData struct { // 东财json
	Data struct {
		Data []struct {
			Time    int     `json:"t"`
			Price   float32 `json:"p"`
			Vol     float32 `json:"v"`
			Type    int     `json:"bs"`
			Amount  float32
			Zhudong float32
		} `json:"data"`
	} `json:"data"`
}

// GetDetailStock /* 获取单只股票所有信息 （图表） */
func GetDetailStock(code string) interface{} {
	//最后一位
	var market = "1"
	if code[len(code)-1] == 'Z' {
		market = "0"
	}
	url := URL + "&code=" + code[0:6] + "&market=" + market
	res, err := http.Get(url)
	// 捕获异常
	if err != nil {
		panic(err)
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	// json解析数据
	info := &SourceData{}
	err = json.Unmarshal(body, info)

	//实时分笔数据（12条）
	fenbi := make([]map[string]interface{}, 12)
	length := len(info.Data.Data)

	for i := range fenbi {
		p := &info.Data.Data[length-12+i]
		fenbi[i] = map[string]interface{}{"time": p.Time, "price": p.Price / 1000, "vol": p.Vol, "type": p.Type}
	}

	allLength := 0
	// 数据处理
	for i := 1; i < len(info.Data.Data); i++ {
		// 定义指针
		p := &info.Data.Data[i]
		pLast := &info.Data.Data[i-1]

		p.Price /= 1000
		p.Time /= 100
		p.Amount = p.Price * p.Vol //成交额
		p.Zhudong = p.Amount       //主动资金

		if p.Type == 1 {
			p.Zhudong *= -1 //主动卖盘 * -1
		} else if p.Type == 4 {
			p.Zhudong *= 0 // 中性盘 * 0
		}
		//主动资金累加
		p.Zhudong += pLast.Zhudong
		//分钟内成交量累加
		if pLast.Time == p.Time {
			p.Vol += pLast.Vol
			p.Amount += pLast.Amount
		}
		if p.Time <= 930 {
			continue
		}
		// 计算最终数组长度
		if p.Time != pLast.Time {
			allLength++
		}
	}
	allLength++
	// 创建数组
	times := make([]int, allLength)
	price := make([]float32, allLength)
	vol := make([]float32, allLength)
	amount := make([]float32, allLength)
	avg := make([]float32, allLength)
	zhudong := make([]float32, allLength)
	//添加数据
	index := 0
	for i := 1; i < len(info.Data.Data); i++ {
		// 定义指针
		p := &info.Data.Data[i]
		pLast := &info.Data.Data[i-1]
		if p.Time <= 930 {
			continue
		}
		// 添加分钟内最后一条数据
		if p.Time != pLast.Time {
			times[index] = pLast.Time
			price[index] = pLast.Price
			vol[index] = pLast.Vol
			amount[index] = pLast.Amount
			zhudong[index] = pLast.Zhudong
			index++
		}
	}
	//添加最后一条数据
	temp := len(info.Data.Data) - 1
	p := &info.Data.Data[temp]
	times[index] = p.Time
	price[index] = p.Price
	vol[index] = p.Vol
	zhudong[index] = p.Zhudong

	var sumAmount float32
	var sumVol float32
	var i int
	for i = range vol {
		sumAmount += amount[i]
		sumVol += vol[i]
		avg[i] = sumAmount / sumVol
	}
	avg = append(avg, avg[i])

	// 字典类型
	mapData := map[string]interface{}{
		"chart": map[string]interface{}{
			"times": times, "price": price, "vol": vol, "avg": avg, "zhudong": zhudong,
		},
		"ticks": fenbi,
		"items": GetSimpleStock([]string{code}),
	}
	return mapData
}

// GetSimpleStock /* 获取多只股票简略信息 */
func GetSimpleStock(codes []string) []map[string]interface{} {
	// 设置最大数量
	if len(codes) > 30 {
		return []map[string]interface{}{}
	}
	// 初始化
	results := make([]map[string]interface{}, 0)

	for _, code := range codes {
		flag := 0
		// 从股票中搜索
		for _, item := range download.AllStock {
			if item["code"] == code {
				results = append(results, item)
				flag = 1
				break
			}
		}
		// 为指数
		if flag == 0 {
			for _, item := range download.AllIndex {
				if item["code"] == code {
					results = append(results, item)
					break
				}
			}
		}
	}
	return results
}

// Search /* 搜索股票 */
func Search(input string, searchType string) []map[string]interface{} {
	// 搜索目标
	temp := download.AllStock
	if searchType == "index" {
		temp = download.AllIndex
	}

	results := make([]map[string]interface{}, 0)
	for _, item := range temp {
		// 搜索前10只
		if len(results) > 10 {
			break
		}
		// 匹配字符串
		res := strings.Contains(item["code"].(string)+item["name"].(string), input)
		if res {
			maps := map[string]interface{}{
				"code": item["code"], "name": item["name"], "price": item["price"], "pct_chg": item["pct_chg"],
			}
			results = append(results, maps)
		}
	}
	return results
}
