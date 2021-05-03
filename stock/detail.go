package stock

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	test "test/redis"
)

const (
	URL = "https://push2ex.eastmoney.com/getStockFenShi?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wzfscj&pageindex=0&sort=1&ft=1"
)

type SourceData struct { // 东财json
	Data struct {
		Data []struct {
			Time            int     `json:"t"`
			Price           float32 `json:"p"`
			Vol             float32 `json:"v"`
			Type            int     `json:"bs"`
			Amount, Zhudong float32
		} `json:"data"`
	} `json:"data"`
}

func GetDetail(code string) interface{} { // 获取股票详情接口
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
			avg[index] = pLast.Amount / pLast.Vol
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
	avg[index] = p.Amount / p.Vol
	zhudong[index] = p.Zhudong

	// 字典类型
	mapData := map[string]interface{}{
		"chart": map[string]interface{}{
			"times": times, "price": price, "vol": vol, "avg": avg, "zhudong": zhudong,
		},
		"ticks": fenbi,
		"items": test.GetDetailStock(code),
	}
	return mapData
}
