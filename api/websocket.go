package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"test/marketime"
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

func Detail(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议失败，", err)
	}
	defer ws.Close()
	//设置超时
	err = ws.SetReadDeadline(time.Now().Add(OVERTIME))
	//读取客户端发送的数据
	mt, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println("读取数据失败，", err)
		return
	}
	code := string(msg)
	// code个数超过1
	if len(strings.Split(code, ",")) > 1 {
		err = ws.WriteMessage(mt, []byte("代码数量不能超过1！"))
		return
	}
	Vol := stock.GetDetailStocks([]string{code})[0]["vol"]
	//写入ws数据
	err = ws.WriteJSON(stock.GetDetailData(code))

	for newVol := stock.GetDetailStocks([]string{code})[0]["vol"]; marketime.IsOpen(); {
		if newVol == Vol {
			time.Sleep(time.Millisecond * 300)
			continue
		}
		Vol = newVol
		//写入ws数据
		err = ws.WriteJSON(stock.GetDetailData(code))
	}
}

func Simple(c *gin.Context) {
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

	for newData := stock.GetSimpleStocks(codes); marketime.IsOpen(); {
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
