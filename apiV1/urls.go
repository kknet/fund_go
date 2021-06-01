package apiV1

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"test/common"
	"test/stock"
)

// GetChart 获取图表数据
func GetChart(c *gin.Context) {
	code := c.Query("code")
	//data := stock.GetSimpleMinute(code)
	data := stock.GetMinuteChart(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// GetStockList 获取股票列表
func GetStockList(c *gin.Context) {
	var data []bson.M
	// 指定code
	codes := c.Query("code")
	if codes != "" {
		clist := strings.Split(codes, ",")
		data = stock.GetStockList(clist)
	}
	// 可指定chart, 获取简略图表数据
	switch c.Query("chart") {
	case "minute", "trends":
		data = common.GoFunc(data, stock.AddSimpleMinute)
	}
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

func GetRank(c *gin.Context) {

}

func Search(c *gin.Context) {
	input := c.Query("input")
	marketType := c.Query("marketType")
	data := stock.Search(input, marketType)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// GetMarket 市场页面聚合接口
func GetMarket(c *gin.Context) {
	marketType := c.Query("marketType")
	c.JSON(200, gin.H{
		"status": true, "data": bson.M{
			"numbers":  stock.GetNumbers(marketType),
			"industry": stock.GetIndustry(marketType),
		},
	})
}

func GetTicks(c *gin.Context) {
	code := c.Query("code")
	data := stock.GetRealtimeTicks(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

func GetPanKou(c *gin.Context) {
	code := c.Query("code")
	data := stock.GetPanKou(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}
