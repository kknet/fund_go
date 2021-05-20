package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"test/apiV1"
	"test/download"
)

/* 主函数 */
func main() {
	// 启动后台下载
	download.GoDownload()

	// 设置日志
	gin.DisableConsoleColor()
	f, err := os.Create("./logs/run.log")
	if err != nil {
		log.Println("Could not open log.")
		panic(err.Error())
	}
	gin.DefaultWriter = io.MultiWriter(f)

	gin.SetMode(gin.ReleaseMode)
	// 创建实例
	r := gin.Default()

	// ApiV1
	v1 := r.Group("/api/v1")
	ws := r.Group("/ws")

	// Stock
	v1Stock := v1.Group("/stock")
	v1Stock.GET("/clist", apiV1.GetStockList)
	v1Stock.GET("/chart", apiV1.GetChart)
	v1Stock.GET("/market", apiV1.GetMarket)
	v1Stock.GET("/ticks", apiV1.GetTicks)

	// WebSocket
	wsStock := ws.Group("/stock")
	wsStock.GET("/detail", apiV1.Detail)
	wsStock.GET("/simple", apiV1.Simple)

	// 错误处理
	r.NoRoute(func(context *gin.Context) {
		context.JSON(http.StatusNotFound, gin.H{"Status": 404, "msg": "Page Not Found"})
	})

	err = r.Run("localhost:10888")
	if err != nil {
		panic(err)
	}
}
