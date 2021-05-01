package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"test/stock"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func (r *http.Request) bool {
		return true
	},
}

func ping(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("err1", err)
	}
	defer ws.Close()
	for {
		//读取ws中的数据
		mt, message, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("err2", err)
			break
		}
		code := string(message)
		fmt.Println(code)
		msg := stock.GetDetail()
		//写入ws数据
		err = ws.WriteMessage(mt, msg)
		if err != nil {
			fmt.Println("err3", err)
			break
		}
	}
}

func main() {
	r := gin.Default()
	r.GET("/detail", ping)
	r.Run("localhost:8080")
}