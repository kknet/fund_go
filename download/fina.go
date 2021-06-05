package download

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strings"
	"test/common"
)

const (
	XueQiuURL = "https://xueqiu.com/service/v5/stock/screener/screen?"
)

func getFinaData(marketType string) {
	url := XueQiuURL + "size=5000&only_count=0&category=" + marketType
	// 添加参数
	rename := map[string]string{
		"symbol": "code",
		"pettm":  "pe_ttm", "pb": "pb", "roediluted": "roe", "eps": "每股收益", "bps": "每股净资产",
		"npt": "净利润", "bi": "营业收入", "bp": "营业利润", "npay": "净利润同比增长率",
	}
	for i := range rename {
		url += "&" + i
	}
	request := common.NewGetRequest(url)
	body, err := request.Do()
	if err != nil {
		log.Println("下载股票数据发生错误，", err.Error())
	}
	str := json.Get(body, "data", "list").ToString()
	//改名
	for i, item := range rename {
		str = strings.Replace(str, i+"\"", item+"\"", -1)
	}
	// json解析
	var temp []bson.M
	_ = json.Unmarshal([]byte(str), &temp)
	for _, i := range temp {
		i["code"] = i["code"].(string)[2:] + "." + i["code"].(string)[0:2]
		i["_id"] = i["code"]
	}
	writeToFina(temp)
	fmt.Println(marketType + "财务指标更新完成")
}

func GetFina() {
	go getFinaData("CN")
}
