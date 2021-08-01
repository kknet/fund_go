package real

import (
	"fund_go2/common"
	"fund_go2/download"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
)

// GetChart 获取图表数据
func GetChart(c *gin.Context) {
	chartType := c.Param("chart_type")
	switch chartType {
	case "north":
		data := GetNorthFlow()
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	case "main_net":
		data := GetMainNetFlow()
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	default:
		c.JSON(200, gin.H{
			"status": false, "msg": "该页面不存在",
		})
	}
}

// GetCList 获取股票列表
func GetCList(c *gin.Context) {
	var data []bson.M
	// 指定code
	codes, ok := c.GetQuery("code")
	if ok {
		clist := strings.Split(codes, ",")
		data = GetStockList(clist)
	} else {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定code参数",
		})
	}
	// 可指定chart, 获取简略图表数据
	switch c.Query("chart") {
	case "minute", "trends":
		data = common.GoFunc(data, AddSimpleMinute)
	case "60day":
		data = common.GoFunc(data, Add60day)
	}
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// GetRank 获取市场排名
func GetRank(c *gin.Context) {
	var query string
	opt := &common.RankOpt{}

	//marketType
	query, ok := c.GetQuery("marketType")
	if !ok {
		c.JSON(200, gin.H{
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
	opt.Page, _ = strconv.ParseInt(page, 8, 64)

	data := getRank(opt)
	// 可指定chart, 获取简略图表数据
	switch c.Query("chart") {
	case "minute", "trends":
		data = common.GoFunc(data, AddSimpleMinute)
	}
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// Search 搜索全市场股票
func Search(c *gin.Context) {
	input, ok := c.GetQuery("input")
	if ok {
		data := search(input)
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	} else {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定input参数",
		})
	}
}

// GetMarket 市场页面聚合接口
func GetMarket(c *gin.Context) {
	marketType, ok := c.GetQuery("marketType")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定marketType参数",
		})
		return
	}
	if marketType == "CN" {
		var industry, sw, area []bson.M
		options := bson.M{"_id": 0, "fmc": 0, "mc": 0, "pe_ttm": 0, "pb": 0}

		_ = download.RealColl.Find(ctx, bson.M{"type": "industry"}).Select(options).All(&industry)
		_ = download.RealColl.Find(ctx, bson.M{"type": "sw"}).Select(options).All(&sw)
		_ = download.RealColl.Find(ctx, bson.M{"type": "area"}).Select(options).All(&area)

		c.JSON(200, gin.H{
			"status": true, "data": bson.M{
				"numbers":  getNumbers(marketType),
				"industry": industry,
				"sw":       sw,
				"area":     area,
			},
		})
	} else {
		c.JSON(200, gin.H{
			"status": true, "data": bson.M{
				"numbers": getNumbers(marketType),
			},
		})
	}
}

func GetTicks(c *gin.Context) {
	code := c.Query("code")
	data := GetRealTicks(code, 50)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}
