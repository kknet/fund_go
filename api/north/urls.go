package north

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

// TopTen 十大成交股
func TopTen(c *gin.Context) {
	data := GetTopTen()
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

type PeriodOptions struct {
	tradeDate string // 日期 二选一
	period    int    // 阶段 二选一
	orderName string
	order     bool
	page      int
	size      int
}

// PeriodData 阶段统计数据
func PeriodData(c *gin.Context) {
	opt := &PeriodOptions{size: 50}

	periodString, ok := c.GetQuery("period")
	// period
	if ok {
		period, err := strconv.Atoi(periodString)
		if err != nil {
			c.JSON(200, gin.H{
				"status": false, "msg": "period参数错误",
			})
		}
		opt.period = period
		opt.orderName = c.DefaultQuery("sort", "values_change")

		// trade_date
	} else {
		opt.tradeDate = c.DefaultQuery("trade_date", "20210727")
		opt.orderName = c.DefaultQuery("sort", "values")
	}

	switch c.DefaultQuery("sorted", "false") {
	case "1", "True", "true":
		opt.order = true
	case "-1", "False", "false":
		opt.order = false
	}

	data := GetPeriodData(opt)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}
