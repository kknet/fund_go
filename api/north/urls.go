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
	period    int
	orderName string
	order     bool
	page      int
	size      int
}

// PeriodData 阶段统计数据
func PeriodData(c *gin.Context) {
	periodString, ok := c.GetQuery("period")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "必须指定period参数",
		})
	}
	period, err := strconv.Atoi(periodString)
	if err != nil {
		c.JSON(200, gin.H{
			"status": false, "msg": "period参数错误",
		})
	}

	opt := &PeriodOptions{
		period:    period,
		orderName: c.DefaultQuery("order_name", "values_change"),
		size:      30,
	}
	order := c.DefaultQuery("order", "true")
	switch order {
	case "1", "True", "true":
		opt.order = true
	case "-1", "False", "false":
		opt.order = false
	default:
		opt.order = false
	}

	data := GetPeriodData(opt)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}
