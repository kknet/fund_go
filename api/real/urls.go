package real

import (
	"fund_go2/common"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
)

// GetChart 获取图表数据
func GetChart(c *gin.Context) {
	code := c.Query("code")
	//data := stock.GetSimpleMinute(code)
	data := GetMinuteData(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// GetCList 获取股票列表
func GetCList(c *gin.Context) {
	var data []map[string]interface{}
	// 指定code
	codes := c.Query("code")
	if codes != "" {
		clist := strings.Split(codes, ",")
		data = GetStockList(clist)
	}
	// 可指定chart, 获取简略图表数据
	switch c.Query("chart") {
	case "minute", "trends":
		data = common.GoFunc(data, AddSimpleMinute)
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
	opt.Page, _ = strconv.Atoi(page)

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

func Search(c *gin.Context) {
	input := c.Query("input")
	data := search(input)
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
				"numbers":  getNumbers(marketType),
				"industry": getIndustry("industry"),
				"sw":       getIndustry("sw"),
				"area":     getIndustry("area"),
			},
		})
	} else {
		c.JSON(200, gin.H{
			"status": true, "data": bson.M{
				"numbers":  getNumbers(marketType),
				"industry": nil,
			},
		})
	}
}

func GetTicks(c *gin.Context) {
	code := c.Query("code")
	data, err := GetRealtimeTicks(code)
	if err != nil {
		c.JSON(200, gin.H{
			"status": false, "msg": err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	}
}

func GetPanKou(c *gin.Context) {
	code := c.Query("code")
	data := PanKou(code)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

// FilterStock 指标选股
func FilterStock(c *gin.Context) {
	marketType, ok := c.Get("marketType")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "请指定marketType参数",
		})
		return
	}
	c.JSON(200, gin.H{
		"status": true, "msg": marketType,
	})
}
