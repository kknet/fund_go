package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"test/api"
	"test/download"
	"test/stock"
)

func main() {
	// 下载
	go download.GetStock(1)
	go download.GetStock(2)
	go download.GetStock(3)

	r := gin.Default()

	// websocket专用
	r.GET("/ws/stock/detail", api.Detail)
	r.GET("/ws/stock/simple", api.Simple)

	// http请求
	r.GET("/api/v1/stock/detail", func(c *gin.Context) {
		str := c.Param("code")
		codes := strings.Split(str, ",")
		data := stock.GetSimpleStocks(codes)
		fmt.Println(data)
		c.JSON(200, gin.H{
			"data": data,
		})
	})

	err := r.Run("localhost:10888")
	if err != nil {
		panic(err)
	}
}
