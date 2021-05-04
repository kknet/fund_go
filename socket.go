package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"test/download"
	"test/stock"
	"time"
)

const (
	OVERTIME = time.Second * 3 // 连接超时
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func IsOpen() bool { //是否开市
	// 工作日
	if time.Now().Weekday() < 5 {
		// 上午
		if time.Now().Hour() >= 9 && time.Now().Hour() < 15 {
			return true
		}
	}
	return false
}

func detail(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议失败，", err)
	}
	defer ws.Close()
	//设置超时
	err = ws.SetReadDeadline(time.Now().Add(OVERTIME))
	//读取客户端发送的数据
	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println("读取数据失败，", err)
		return
	}
	// codes := strings.Split(string(msg), ",")

	oldData := stock.GetDetailData(string(msg))
	//写入ws数据
	err = ws.WriteJSON(oldData)
}

func simple(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议失败，", err)
	}
	defer ws.Close()
	//设置超时
	err = ws.SetReadDeadline(time.Now().Add(OVERTIME))
	//读取客户端发送的数据
	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println("读取数据失败，", err)
		return
	}
	codes := strings.Split(string(msg), ",")

	oldData := stock.GetSimpleStocks(codes)
	//写入ws数据
	err = ws.WriteJSON(oldData)

	for newData := stock.GetSimpleStocks(codes); IsOpen(); {
		flag := false

		for i := range newData {
			if oldData[i]["price"] != newData[i]["price"] {
				// 写入新数据
				oldData[i]["price"] = newData[i]["price"]
				flag = true
			}
		}
		if flag {
			//写入ws数据
			err = ws.WriteJSON(oldData)
		}
	}
}

func main() {
	// 下载
	go download.GetStock(1)
	go download.GetStock(2)
	go download.GetStock(3)

	r := gin.Default()
	// websocket专用
	r.GET("/ws/stock/detail", detail)
	r.GET("/ws/stock/simple", simple)
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
	err := r.Run("localhost:8080")
	if err != nil {
		panic(err)
	}
}
