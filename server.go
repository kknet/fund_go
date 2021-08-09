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

	// ApiV1
	v1 := r.Group("/api/v1")
	ws := r.Group("/ws")

	// WebSocket
	wsStock := ws.Group("/stock")
	wsStock.GET("/clist", api.ConnectCList)
	wsStock.GET("/items", api.ConnectItems)

	// Real 实时数据
	Real := v1.Group("/stock")
	Real.GET("/chart/:chart_type", real.GetChart)
	Real.GET("/market", real.GetMarket)
	Real.GET("/ticks", real.GetTicks)

	// CList
	CList := Real.Group("/clist")
	CList.GET("/get", real.GetCList)
	CList.GET("/rank", real.GetRank)
	CList.GET("/search", real.Search)
	CList.GET("/members", real.GetMembers)

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
