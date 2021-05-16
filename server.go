package main

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"test/api"
	"test/download"
	"test/stock"
)

const ( // url前缀
	WsUrl  = "/ws"
	ApiUrl = "/api/v1"
)

/* 主函数 */
func main() {
	// 下载
	download.GoDownload()

	r := gin.Default()

	// websocket专用
	r.GET(WsUrl+"/stock/detail", api.Detail)
	r.GET(WsUrl+"/stock/simple", api.Simple)
	//r.GET(WsUrl+"/stock/rank", api.Rank)

	// http请求
	r.GET(ApiUrl+"/stock/detail", func(c *gin.Context) {
		code := c.Query("code")
		data := stock.GetDetailData(code)
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	})

	r.GET(ApiUrl+"/stock/simple", func(c *gin.Context) {
		code := c.Query("code")
		codes := strings.Split(code, ",")
		data := stock.GetStockList(codes)
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	})

	// 市场页面聚合接口
	r.GET(ApiUrl+"/stock/market", func(c *gin.Context) {
		marketType := c.Query("marketType")
		c.JSON(200, gin.H{
			"status": true, "data": bson.M{
				"numbers":  stock.GetNumbers(marketType),
				"industry": stock.GetIndustry(marketType),
			},
		})
	})

	r.GET(ApiUrl+"/stock/search", func(c *gin.Context) {
		input := c.Query("input")
		searchType := c.Query("type")
		data := stock.Search(input, searchType)
		c.JSON(200, gin.H{
			"status": true, "data": data,
		})
	})

	err := r.Run("localhost:10888")
	if err != nil {
		panic(err)
	}
}
