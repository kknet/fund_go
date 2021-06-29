package fina

import "github.com/gin-gonic/gin"

func GetFina(c *gin.Context) {
	code := c.Query("code")
	data := GetFinaData(code)
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
