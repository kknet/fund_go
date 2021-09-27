package real

import (
	"fund_go2/common"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ControlLimitFunc 流量控制中间件
func ControlLimitFunc(c *gin.Context) {
	ip, ok := c.RemoteIP()
	if !ok {
		return
	}
	// 是否存在
	exists, _ := limitDB.Exists(ctx, ip.String()).Result()
	if exists >= 1 {
		times, _ := limitDB.Incr(ctx, ip.String()).Result()
		// 半小时之内 访问次数超过4000 则拒绝
		if times > 4000 {
			// forbidden
			c.JSON(http.StatusForbidden, bson.M{
				"status": false, "msg": "请求被拦截",
			})
			c.Abort()
		}
	} else {
		limitDB.SetEX(ctx, ip.String(), 1, time.Minute*30)
	}
	c.Next()
}

// CheckCodeFunc 检查代码中间件
func CheckCodeFunc(c *gin.Context) {
	_, ok := c.GetQuery("code")
	if !ok {
		c.JSON(http.StatusOK, bson.M{
			"status": false, "msg": "必须指定code参数",
		})
		c.Abort()
	}
	c.Next()
}

// GetChart 获取图表数据
func GetChart(c *gin.Context) {
	chartType := c.Param("chart_type")
	code := c.Query("code")

	switch chartType {
	// 资金博弈
	case "zjby":
		data := GetDetailMoneyFlow(code)
		c.JSON(http.StatusOK, gin.H{
			"status": true, "data": data,
		})
	case "minute":
		data := GetIndustryMinute(code)
		c.JSON(http.StatusOK, gin.H{
			"status": true, "data": data,
		})
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"status": false, "msg": "该页面不存在",
		})
	}
}

// StockDetail 获取单只股票详细数据
func StockDetail(c *gin.Context) {
	code := c.Query("code")
	data := GetStock(code, true)
	c.JSON(http.StatusOK, gin.H{
		"status": true, "data": data,
	})
}

// StockList 获取股票列表
func StockList(c *gin.Context) {
	code := c.Query("code")
	codeList := strings.Split(code, ",")
	data := GetStockList(codeList)

	// 可指定chart, 获取简略图表数据
	switch c.Query("chart") {
	case "minute", "trends":
		data = common.GoFunc(data, AddSimpleMinute)
	case "60day":
		data = common.GoFunc(data, Add60day)
	case "main_net":
		data = common.GoFunc(data, AddMainFlow)
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
	opt.Page, _ = strconv.ParseInt(page, 16, 64)

	data := getRank(opt)
	// 可指定chart, 获取简略图表数据
	switch c.Query("chart") {
	case "minute", "trends":
		data = common.GoFunc(data, AddSimpleMinute)
	case "60day":
		data = common.GoFunc(data, Add60day)
	case "main_net":
		data = common.GoFunc(data, AddMainFlow)
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

		c.JSON(200, gin.H{
			"status": true, "data": bson.M{
				"numbers":  getNumbers(marketType),
				"industry": GetSimpleBK("industry"),
				"concept":  GetSimpleBK("concept"),
				"area":     GetSimpleBK("area"),
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

// ViewItemsPage 详情页面汇总数据
func ViewItemsPage(c *gin.Context) {
	code := c.Query("code")
	items := GetStock(code, true)
	if len(items) == 0 {
		c.JSON(200, gin.H{
			"status": false, "msg": "code不存在",
		})
		return
	}
	pankou, ticks := GetRealTicks(items)
	view := viewPage(code)

	c.JSON(200, gin.H{
		"status": true, "data": bson.M{
			"items": items, "ticks": ticks, "pankou": pankou, "view": view,
		},
	})
}
