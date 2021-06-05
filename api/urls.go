package apiV1

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"test/api/stock"
	"test/common"
)

// GetChart 获取图表数据
func GetChart(c *gin.Context) {
	//code := c.Query("code")
	//data := stock.GetSimpleMinute(code)
	//data := stock.GetMinuteChart(code)
	c.JSON(200, gin.H{
		"status": true, "msg": "该接口暂不可用",
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

// GetRank 获取市场排名
func GetRank(c *gin.Context) {
	var query string
	opt := &common.RankOpt{}
	//获取参数
	//marketType
	query, status := c.GetQuery("marketType")
	if !status {
		c.JSON(400, gin.H{
			"status": false, "msg": "必须指定marketType参数",
		})
		return
	}
	opt.MarketType = query
	//sort
	opt.SortName = c.DefaultQuery("sort", "amount")
	query = c.DefaultQuery("sorted", "false")
	switch query {
	case "1", "true", "True":
		opt.Sorted = true
	default:
		opt.Sorted = false
	}
	//page
	page := c.DefaultQuery("page", "1")
	opt.Page, _ = strconv.ParseInt(page, 10, 64)

	data := stock.GetRank(opt)
	// 可指定chart, 获取简略图表数据
	switch c.Query("chart") {
	case "minute", "trends":
		data = common.GoFunc(data, stock.AddSimpleMinute)
	}
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
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
	if marketType == "CN" {
		c.JSON(200, gin.H{
			"status": true, "data": bson.M{
				"numbers":  stock.GetNumbers(marketType),
				"industry": stock.GetIndustry("industry"),
				"sw":       stock.GetIndustry("sw"),
				"area":     stock.GetIndustry("area"),
			},
		})
	} else {
		c.JSON(200, gin.H{
			"status": true, "data": bson.M{
				"numbers": stock.GetNumbers(marketType),
			},
		})
	}
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
