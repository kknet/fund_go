package stock

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// 东财json

type SourceData struct {
	Data struct {
		Data []struct {
			Time    int     `json:"t"`
			Price   float32 `json:"p"`
			Vol     float32 `json:"v"`
			Amount  float32
			Type    int `json:"bs"`
			Zhudong float32
		} `json:"data"`
	} `json:"data"`
}

// 返回json

type MyData struct {
	Time    int
	Price   float32
	Vol     float32
	Avg     float32
	Zhudong float32
}

// 获取股票详情接口

func GetDetail() []byte {
	// 数据来源 东方财富网
	url1 := "https://push2ex.eastmoney.com/getStockFenShi?ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wzfscj"
	url2 := "&pageindex=0&sort=1&ft=1&code=002714&market=0"
	res, err := http.Get(url1 + url2)
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

	// 数据处理
	for i := range info.Data.Data {
		// 定义指针
		p := &info.Data.Data[i]

		p.Price /= 1000
		p.Time /= 100
		p.Amount = p.Price * p.Vol //成交额
		p.Zhudong = p.Amount       //主动资金

		if p.Type == 1 {
			p.Zhudong *= -1 //主动卖盘 * -1
		} else if p.Type == 4 {
			p.Zhudong *= 0 // 中性盘 * 0
		}

		if i >= 1 {
			//上条数据的指针
			pLast := &info.Data.Data[i-1]
			//主动资金累加
			p.Zhudong += pLast.Zhudong
			//分钟内成交量累加
			if pLast.Time == p.Time {
				p.Vol += pLast.Vol
				p.Amount += pLast.Amount
			}
		}
	}

	mySlice := make([]MyData, 0, 0)
	//添加数据
	for i := 1; i < len(info.Data.Data); i++ {
		// 定义指针
		p := &info.Data.Data[i]
		pLast := &info.Data.Data[i-1]
		if p.Time <= 930 {
			continue
		}

		if p.Time != pLast.Time {
			mySlice = append(mySlice, MyData{
				Time: pLast.Time, Vol: pLast.Vol, Price: pLast.Price,
				Zhudong: pLast.Zhudong, Avg: pLast.Amount / pLast.Vol,
			})
		}
	}
	//添加最后一条数据
	p := &info.Data.Data[len(info.Data.Data)-1]
	mySlice = append(mySlice, MyData{Time: p.Time, Vol: p.Vol, Price: p.Price, Zhudong: p.Zhudong, Avg: p.Amount / p.Vol})
	// 转换成json格式
	jsonData, err := json.Marshal(mySlice)
	return jsonData
}
