package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"test/api"
	"test/download"
	"test/stock"
)

const ( // url前缀
	WsUrl  = "/ws"
	ApiUrl = "/api/v1"
)

func main() {
	// 下载
	download.GoDownload()

	r := gin.Default()

	// websocket专用
	r.GET(WsUrl+"/stock/detail", api.Detail)
	r.GET(WsUrl+"/stock/simple", api.Simple)
	r.GET(WsUrl+"/stock/rank", api.Rank)

	// http请求
	r.GET(ApiUrl+"/stock/detail", func(c *gin.Context) {
		code := c.Query("code")
		data := stock.GetDetailData(code)
		c.JSON(200, gin.H{
			"data": data,
		})
	})

	r.GET(ApiUrl+"/stock/simple", func(c *gin.Context) {
		code := c.Query("code")
		codes := strings.Split(code, ",")
		data := stock.GetSimpleStocks(codes)
		c.JSON(200, gin.H{
			"data": data,
		})
	})

	err := r.Run("localhost:10888")
	if err != nil {
		panic(err)
	}
	fmt.Println("服务启动成功！")
}
