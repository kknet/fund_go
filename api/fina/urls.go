package fina

import "github.com/gin-gonic/gin"

func GetFina(c *gin.Context) {
	code, ok := c.GetQuery("code")
	if !ok {
		c.JSON(200, gin.H{
			"status": false, "msg": "未指定code参数",
		})
		return
	}
	period := c.DefaultQuery("period", "y")
	data := GetFinaData(code, period)
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}

func Filter(c *gin.Context) {
	data := FilterStock()
	c.JSON(200, gin.H{
		"status": true, "data": data,
	})
}
