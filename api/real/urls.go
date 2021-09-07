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
	case "detail_money":
		code, ok := c.GetQuery("code")
		if !ok {
			c.JSON(200, gin.H{
				"status": false, "msg": "必须指定code参数",
			})
			return
		}
		data := GetDetailMoneyFlow(code)
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	default:
		c.JSON(200, gin.H{
			"status": false, "msg": "该页面不存在",
		})
	}
}

// StockDetail 获取单只股票详细数据
func StockDetail(c *gin.Context) {
	// 指定code
	code, ok := c.GetQuery("code")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定code参数",
		})
		return
	}
	data := GetStock(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// StockList 获取股票列表
func StockList(c *gin.Context) {
	// 指定code
	code, ok := c.GetQuery("code")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定code参数",
		})
		return
	}
	codeList := strings.Split(code, ",")
	data := GetStockList(codeList)

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
		options := bson.M{"_id": 0, "code": 1, "name": 1, "pct_chg": 1, "领涨股": 1, "max_pct": 1, "main_net": 1}

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

// GetMembers 获取成分股
func GetMembers(c *gin.Context) {
	code, ok := c.GetQuery("code")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定code参数",
		})
	}
	data := GetIndustryMembers(code)
	data = common.GoFunc(data, AddSimpleMinute)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// GetTicks 获取五档盘口、成交明细
func GetTicks(c *gin.Context) {
	code, ok := c.GetQuery("code")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定code参数",
		})
	}
	data := GetRealTicks(code, 50)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}
