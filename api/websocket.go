package apiV1

import (
	"fund_go2/api/real"
	"fund_go2/download"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	OVERTIME = time.Second * 5 // 连接超时
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 建立连接
func connect(c *gin.Context) *websocket.Conn {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if ws == nil {
		return nil
	}
	if err != nil {
		_ = ws.WriteJSON(bson.M{"msg": "连接服务器错误", "status": false})
		return nil
	}
	return ws
}

// ConnectCList 自选表连接
func ConnectCList(c *gin.Context) {
	ws := connect(c)
	//设置超时
	err := ws.SetReadDeadline(time.Now().Add(OVERTIME))
	//读取代码
	_, msg, err := ws.ReadMessage()
	if err != nil {
		_ = ws.WriteJSON(bson.M{"msg": "读取代码超时", "status": false})
		return
	}
	codes := strings.Split(string(msg), ",")
	AddToConnList(ws, codes, "clist")
}

// ConnectItems 详情页连接
func ConnectItems(c *gin.Context) {
	ws := connect(c)
	//设置超时
	err := ws.SetReadDeadline(time.Now().Add(OVERTIME))
	//读取代码
	_, msg, err := ws.ReadMessage()
	if err != nil {
		_ = ws.WriteJSON(bson.M{"msg": "读取代码超时", "status": false})
		return
	}
	// 检查代码
	codes := []string{string(msg)}
	check := real.GetStockList(codes)
	if len(check) > 0 {
		AddToConnList(ws, codes, "items")
	} else {
		_ = ws.WriteJSON(bson.M{"msg": "代码不存在", "status": false})
	}
}

// SendCList 推送消息
func SendCList() {
	for _, c := range ConnList["clist"] {
		err := c.Ping()
		if err != nil {
			Close(c)
			continue
		}
		// 获取新数据
		newData := real.GetStockList(c.codes)
		for i := range newData {
			if newData[i]["pct_chg"] != c.data[i]["pct_chg"] {
				c.data[i] = newData[i]
				real.AddSimpleMinute(newData[i])
				// 写入
				err = c.Conn.WriteJSON(newData[i])
				// 错误 关闭连接
				if err != nil {
					Close(c)
					break
				}
			}
		}
	}
}

// SendItems 推送详情页
func SendItems() {
	for _, c := range ConnList["items"] {
		err := c.Ping()
		if err != nil {
			Close(c)
			continue
		}
		// 获取新数据
		newData := real.GetStockList(c.codes)[0]

		results := make(map[string]interface{})

		// 有更新
		if newData["vol"] != c.data[0]["vol"] {
			c.data[0] = newData

			group := sync.WaitGroup{}
			group.Add(3)

			// 详情
			results["items"] = newData
			// 盘口明细
			go func() {
				if newData["marketType"] == "CN" {
					results["pankou"] = real.PanKou(c.codes[0])
				}
				group.Done()
			}()
			// 实时分笔
			go func() {
				results["ticks"], _ = real.GetRealtimeTicks(c.codes[0])
				group.Done()
			}()
			// 分时图
			go func() {
				if newData["pct_chg"] != c.data[0]["pct_chg"] {
					results["minute"] = real.GetMinuteData(c.codes[0])
				}
				group.Done()
			}()
			group.Wait()

			c.data[0] = newData
			// 写入
			err = c.Conn.WriteJSON(results)
			// 错误 关闭连接
			if err != nil {
				Close(c)
				break
			}
		}
	}
}

// ListenChan 总监听函数
func ListenChan() {
	for {
		<-download.MyChan
		SendItems()
		SendCList()
	}
}

// Close 关闭连接
func Close(conn *MyConn) {
	for i, c := range ConnList[conn.Type] {
		if c.uid == conn.uid {
			_ = conn.Conn.Close()
			ConnList[conn.Type] = append(ConnList[conn.Type][:i], ConnList[conn.Type][i+1:]...)
			break
		}
	}
}
