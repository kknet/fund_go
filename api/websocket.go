package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"reflect"
	"strings"
	"test/download"
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

// Detail /* 实时图表行情 */
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
	oldData := stock.GetStockList([]string{code})
	//写入ws数据
	err = ws.WriteJSON(stock.GetDetailData(code))

	for marketime.IsOpen() {
		// 阻塞
		_ = <-download.CNChan

		newData := stock.GetStockList([]string{code})
		// 相等则不写入，继续阻塞
		if reflect.DeepEqual(oldData, newData) {
			continue
		}
		oldData = newData
		//写入ws数据
		err = ws.WriteJSON(stock.GetDetailData(code))
	}
}

// Simple /* 简略行情 */
func Simple(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议失败，", err)
	}
	// 函数退出时关闭ws
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

	oldData := stock.GetStockList(codes)
	//写入ws数据
	err = ws.WriteJSON(oldData)

	for marketime.IsOpen() {
		// 阻塞
		_ = <-download.CNChan
		//获取新数据
		newData := stock.GetStockList(codes)
		// 相等则不写入，继续阻塞
		if reflect.DeepEqual(oldData, newData) {
			continue
		}
		// 写入新数据
		err = ws.WriteJSON(oldData)
		oldData = newData
	}
}

/*
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
	//添加简略行情数据
	oldData := stock.GetStocks(codes)

	// 与排行榜数据合并
	for i := range oldData {
		oldData[i][rankName] = scores[codes[i]]
	}
	//写入ws数据
	err = ws.WriteJSON(oldData)
	if err != nil {
		log.Println(err)
	}

	for marketime.IsOpen() {
		time.Sleep(time.Second * 3)
		// 获取排行榜
		scores := stock.GetRank(rankName)
		codes := []string{}
		// 获取代码
		for i := range scores {
			codes = append(codes, i)
		}
		//添加简略行情数据
		oldData := stock.GetStocks(codes)
		// 与排行榜数据合并
		for i := range oldData {
			oldData[i][rankName] = scores[codes[i]]
		}
		//写入ws数据
		err = ws.WriteJSON(oldData)
	}
}
*/
