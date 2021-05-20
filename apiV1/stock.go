package apiV1

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"test/stock"
)

func GetDetail(c *gin.Context) {
	code := c.Query("code")
	data := stock.GetDetailData(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// GetChart 获取图表数据
func GetChart(c *gin.Context) {
	code := c.Query("code")
	data := stock.GetMinuteChart(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// GetStockList 获取股票列表
// 以下为几种不同的获取方式
// 1. 指定code，如：code=000001.SH, 600519.SH, 00700.HK, AAPL.US
// 2. 指定search搜索，如search="贵州" 或 “60051” 搜索到贵州茅台
// 3. 指定size, page, sort可获取排行榜，如size=10, page=2, sort="vol" 获取成交量在全市场10-20名的股票
func GetStockList(c *gin.Context) {

	opt := stock.CListOpt{
		Codes:      strings.Split(c.Query("code"), ","),
		MarketType: c.DefaultQuery("marketType", "CN"),
		Search:     c.Query("search"),
		SortName:   c.Query("sort"),
	}
	switch c.Query("order") {
	case "1", "T", "f", "true", "True":
		opt.Sorted = true
	default:
		opt.Sorted = false
	}
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	opt.Size = size
	opt.Page = page

	data := stock.GetStockList(opt)
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
