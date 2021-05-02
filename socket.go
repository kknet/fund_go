package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
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
	// 获取代码
	code := string(msg)
	// 初始化
	vol, err := test.GetVol(code)
	if err != nil {
		// 写入错误信息
		if err == redis.Nil {
			err = ws.WriteMessage(mt, []byte("该代码不存在！"))
		} else {
			err = ws.WriteMessage(mt, []byte(err.Error()))
			return
		}
	}
	//写入ws数据
	err = ws.WriteMessage(mt, stock.GetDetail(code))
	if err != nil {
		fmt.Println("写入数据失败，", err)
	}
	for {
		// 检查是否需要更新
		nowVol, _ := test.GetVol(code)
		if vol == nowVol {
			fmt.Println("无需更新！")
			// 1s间隔
			time.Sleep(time.Millisecond * 1000)
			continue
		}
		vol = nowVol
		//写入ws数据
		err = ws.WriteMessage(mt, stock.GetDetail(code))
		if err != nil {
			fmt.Println("写入数据失败，", err)
			break
		}
	}
}

func main() {
	r := gin.Default()
	r.GET("/detail", detail)
	r.Run("localhost:8080")
}
