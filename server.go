package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"test/apiV1"
	"test/download"
)

/* 主函数 */
func main() {
	// 启动后台下载
	download.GoDownload()

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

	// ApiV1
	v1 := r.Group("/api/v1")
	ws := r.Group("/ws")

	// Real 实时数据
	Real := v1.Group("/stock")

	// CList
	CList := Real.Group("/clist")
	CList.GET("/get", apiV1.GetStockList)
	CList.GET("/rank", apiV1.GetRank)
	CList.GET("/search", apiV1.Search)

	Real.GET("/chart", apiV1.GetChart)
	Real.GET("/market", apiV1.GetMarket)
	Real.GET("/ticks", apiV1.GetTicks)
	Real.GET("/pankou", apiV1.GetPanKou)

	// WebSocket
	wsStock := ws.Group("/stock")
	wsStock.GET("/items", apiV1.Items)
	wsStock.GET("/simple", apiV1.Simple)

	// 错误处理
	r.NoRoute(func(context *gin.Context) {
		context.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "page not found"})
	})

	// 启动
	err := r.Run("localhost:10888")
	if err != nil {
		panic(err)
	}
}
