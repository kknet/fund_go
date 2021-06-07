package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	apiV1 "test/api"
	"test/api/real"
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
	CList.GET("/get", real.GetCList)
	CList.GET("/rank", real.GetRank)
	CList.GET("/search", real.Search)

	Real.GET("/chart", real.GetChart)
	Real.GET("/market", real.GetMarket)
	Real.GET("/ticks", real.GetTicks)
	Real.GET("/pankou", real.GetPanKou)

	// WebSocket
	wsStock := ws.Group("/stock")
	wsStock.GET("/clist", apiV1.CList)

	// 错误处理
	r.NoRoute(func(context *gin.Context) {
		context.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "未找到该页面"})
	})

	// 启动
	err := r.Run("localhost:10888")
	if err != nil {
		panic(err)
	}
}
