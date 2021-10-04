package main

import (
	api "fund_go2/api"
	"fund_go2/api/fina"
	"fund_go2/api/real"
	"fund_go2/api/user"
	"fund_go2/download"
	"github.com/gin-gonic/gin"
	"net/http"
)

func init() {
	// 监听websocket
	go api.ListenChan()

	// 启动后台下载
	download.GoDownload()
}

func main() {
	//gin.SetMode(gin.ReleaseMode)

	// 创建实例
	r := gin.Default()
	// 全局流量控制
	r.Use(real.ControlLimitFunc)

	// apiV1
	apiV1 := r.Group("/api/v1")
	ws := r.Group("/ws")

	// WebSocket
	wsStock := ws.Group("/stock")
	wsStock.Use(real.CheckCodeFunc)
	wsStock.GET("/list", api.ConnectCList)
	wsStock.GET("/detail", api.ConnectItems)

	// Stock 股票数据
	stock := apiV1.Group("/stock")
	stock.GET("/search", real.Search)
	// 验证代码中间件
	stock.Use(real.CheckCodeFunc)
	stock.GET("/list", real.StockList)
	stock.GET("/detail", real.StockDetail)
	stock.GET("/chart/:chart_type", real.GetChart)
	stock.GET("/page", real.ViewItemsPage)

	// Market 市场数据
	market := apiV1.Group("/market")
	market.GET("/rank", real.GetRank)
	market.GET("/industry", real.GetMarket)
	market.GET("/members", real.GetMembers)

	// Fina 财务数据
	Fina := apiV1.Group("/fina")
	Fina.GET("/get", real.CheckCodeFunc, fina.GetFina)
	Fina.GET("/filter", fina.Filter)

	// User 用户
	User := apiV1.Group("/user")
	// 使用中间件
	User.GET("/info", user.Authorize, user.GetInfo)
	User.POST("/info", user.Register)
	User.PUT("/info", user.Authorize, user.UpdateInfo)
	User.POST("/token", user.Login)
	User.DELETE("/token", user.Authorize, user.Logout)

	// 错误处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "未找到该页面"})
	})

	// 启动
	err := r.Run("0.0.0.0:10888")
	if err != nil {
		panic(err)
	}
}
