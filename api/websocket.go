package api

import (
	"fmt"
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

	for marketime.IsOpen() {
		newVol := stock.GetDetailStocks([]string{code})[0]["vol"]
		if newVol == Vol {
			time.Sleep(time.Millisecond * 100)
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

	for marketime.IsOpen() {
		newData := stock.GetSimpleStocks(codes)

		for i := range newData {
			if oldData[i]["pct_chg"] != newData[i]["pct_chg"] {
				// 数据有变化
				oldData = newData
				// 写入新数据
				err = ws.WriteJSON(oldData)
				break
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func Rank(c *gin.Context) {
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
	rankName := string(msg)

	// 获取排行榜
	scores := stock.GetRank(rankName)
	codes := []string{}
	// 获取代码
	for i := range scores {
		codes = append(codes, i)
	}
	fmt.Println(codes)
	//添加简略行情数据
	tempData := stock.GetSimpleStocks(codes)

	// 与排行榜数据合并
	for i := range tempData {
		tempData[i][rankName] = scores[codes[i]]
	}
	fmt.Println(tempData)
	//写入ws数据
	err = ws.WriteJSON(tempData)
	if err != nil {
		log.Println(err)
	}
	for marketime.IsOpen() {
		time.Sleep(time.Millisecond * 100)
	}
}
