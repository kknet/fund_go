package apiV1

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

func GetStockList(c *gin.Context) {
	code := c.Query("code")
	codes := strings.Split(code, ",")
	data := stock.GetStockList(codes)
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

// Search 搜索
func Search(c *gin.Context) {
	input := c.Query("input")
	searchType := c.Query("type")
	data := stock.Search(input, searchType)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

func Rank(c *gin.Context) {
	marketType := c.Query("marketType")
	data := stock.GetRank(marketType)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}
