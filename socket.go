package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"test/redis"
	"test/stock"
	"time"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func detail(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议失败，", err)
	}
	defer ws.Close()

	//读取客户端发送的数据
	mt, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println("读取数据失败，", err)
	}
	// 获取代码并初始化
	code := string(msg)
	codes := []string{code}
	volMap := test.GetVolMaps(codes)
	if len(volMap) == 0 {
		_ = ws.WriteMessage(mt, []byte("该代码不存在！代码示例：600519.SH"))
		return
	}
	//写入ws数据
	_ = ws.WriteMessage(mt, stock.GetDetail(code))

	// 若正在交易则保持连接
	for 9 <= time.Now().Hour() && time.Now().Hour() < 15 {
		newVolMap := test.GetVolMaps(codes)
		// 检查是否需要更新
		if volMap[code] == newVolMap[code] {
			// 0.1s间隔
			time.Sleep(time.Millisecond * 100)
			continue
		}
		volMap = newVolMap
		//写入ws数据
		_ = ws.WriteMessage(mt, stock.GetDetail(code))
	}
}

func simple(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议失败，", err)
	}
	defer ws.Close()

	//读取客户端发送的数据
	mt, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println("读取数据失败，", err)
	}
	// 根据逗号切片
	codes := strings.Split(string(msg), ",")
	// 初始化
	volMaps := test.GetVolMaps(codes)
	//写入数据
	jsonData, _ := json.Marshal(test.GetSimpleStock(codes))
	_ = ws.WriteMessage(mt, jsonData)
	// 若正在交易则保持连接
	for 9 <= time.Now().Hour() && time.Now().Hour() < 15 {
		newVolMaps := test.GetVolMaps(codes)
		// 检查是否有键值更新
		for i := range volMaps {
			code := volMaps[i]
			// 有更新
			if volMaps[code] != newVolMaps[code] {
				volMaps = newVolMaps
				//写入ws数据
				jsonData, _ := json.Marshal(test.GetSimpleStock(codes))
				_ = ws.WriteMessage(mt, jsonData)
				break
			}
		}
		// 0.1s
		time.Sleep(time.Millisecond * 100)
	}
}

func main() {
	r := gin.Default()
	r.GET("/detail", detail)
	r.GET("/simple", simple)
	err := r.Run("localhost:8080")
	if err != nil {
		panic(err)
	}
}
