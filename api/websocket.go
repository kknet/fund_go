package apiV1

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"test/api/stock"
	"test/download"
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

// Items 详情页面整合行情
func Items(c *gin.Context) {
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
		err = ws.WriteJSON("读取代码超时！")
		return
	}
	code := string(msg)
	// 查看代码是否存在
	single := stock.GetStockList([]string{code})
	_ = ws.WriteJSON(bson.M{"data": single, "type": "items"})
	// 发送信息
	group := sync.WaitGroup{}
	group.Add(2)
	go func() {
		info := stock.GetPanKou(code)
		_ = ws.WriteJSON(bson.M{"data": info, "type": "pankou"})
		group.Done()
	}()
	go func() {
		info := stock.GetRealtimeTicks(code)
		_ = ws.WriteJSON(bson.M{"data": info, "type": "ticks"})
		group.Done()
	}()
	group.Wait()
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

	for {
		// 阻塞
		_ = <-download.MyChan
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

// Ticks 行情
func Ticks(c *gin.Context) {
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

	for {
		// 阻塞
		_ = <-download.MyChan
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
