package apiV1

import (
	"fund_go2/api/real"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
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

// connect 建立连接
func connect(c *gin.Context) *websocket.Conn {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = ws.WriteJSON(bson.M{"msg": "连接服务器错误", "status": false})
		return nil
	}
	return ws
}

// CList 自选页面
func CList(c *gin.Context) {
	ws := connect(c)
	defer ws.Close()
	//设置超时
	err := ws.SetReadDeadline(time.Now().Add(OVERTIME))
	//读取代码
	_, msg, err := ws.ReadMessage()
	if err != nil {
		_ = ws.WriteJSON(bson.M{"msg": "读取数据超时", "status": false})
		return
	}
	codes := strings.Split(string(msg), ",")
	oldData := real.GetStockList(codes)
	for {
		time.Sleep(time.Millisecond * 100)
		//获取新数据
		newData := real.GetStockList(codes)
		//逐个比较
		for i := range oldData {
			for j := range newData {
				// 相同代码
				if oldData[i]["code"] == newData[j]["code"] {
					if oldData[i]["pct_chg"].(float64) != newData[j]["pct_chg"].(float64) {
						temp := newData[j]
						real.AddSimpleMinute(temp)
						//写入
						err = ws.WriteJSON(temp)
						oldData[i] = newData[j]
					}
				}
			}
		}
	}
}
