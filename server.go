package main

import (
	api "fund_go2/api"
	"fund_go2/api/fina"
	"fund_go2/api/real"
	"fund_go2/download"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 主函数
func main() {
	// 监听websocket
	go api.ListenChan()

	// 启动后台下载
	download.GoDownload()

	var err error
	// 设置日志
	//gin.DisableConsoleColor()
	//f, err := os.Create("./logs/run.log")
	//if err != nil {
	//	log.Println("Could not open log.")
	//	panic(err)
	//}
	//gin.DefaultWriter = io.MultiWriter(f)
	//
	//gin.SetMode(gin.ReleaseMode)
	// 创建实例
	r := gin.Default()

	// api
	v1 := r.Group("/api/v1")
	ws := r.Group("/ws")

	// WebSocket
	wsStock := ws.Group("/stock")
	wsStock.GET("/list", api.ConnectCList)
	wsStock.GET("/detail", api.ConnectItems)

	// Stock 股票数据
	stock := v1.Group("/stock")
	stock.GET("/list", real.StockList)
	stock.GET("/detail", real.StockDetail)
	stock.GET("/chart/:chart_type", real.GetChart)
	stock.GET("/ticks", real.GetTicks)
	stock.GET("/search", real.Search)

	// Market 市场数据
	market := v1.Group("/market")
	market.GET("/rank", real.GetRank)
	market.GET("/industry", real.GetMarket)
	market.GET("/members", real.GetMembers)

	// Fina 财务数据
	Fina := v1.Group("/fina")
	Fina.GET("/get", fina.GetFina)
	Fina.GET("/filter", fina.Filter)

	// 错误处理
	r.NoRoute(func(context *gin.Context) {
		context.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "未找到该页面"})
	})

	// 启动
	err = r.Run("0.0.0.0:10888")
	if err != nil {
		panic(err)
	}
}
