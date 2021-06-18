package apiV1

import (
	"fmt"
	"fund_go2/api/real"
	"fund_go2/download"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
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

// Connect 建立连接
func Connect(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if ws == nil {
		return
	}
	if err != nil {
		_ = ws.WriteJSON(bson.M{"msg": "连接服务器错误", "status": false})
		return
	}
	//设置超时
	err = ws.SetReadDeadline(time.Now().Add(OVERTIME))
	//读取代码
	_, msg, err := ws.ReadMessage()
	if err != nil {
		_ = ws.WriteJSON(bson.M{"msg": "读取数据超时", "status": false})
		return
	}
	codes := strings.Split(string(msg), ",")
	uid, _ := uuid.NewUUID()

	conn := &MyConn{
		Conn: ws, codes: codes, uid: uid.String(),
		data: real.GetStockList(codes), CountDown: time.Now(),
	}
	CListConn = append(CListConn, conn)
}

// SendCList 推送消息
func SendCList() {
	for {
		_ = <-download.MyChan

		for _, c := range CListConn {
			// 每隔一分钟检查连接
			if time.Since(c.CountDown) > time.Minute {
				// 写入
				err := c.Conn.WriteJSON(bson.M{"msg": "ping"})
				// 错误 关闭连接
				if err != nil {
					Close(c)
				} else {
					c.CountDown = time.Now()
				}
			}
			// 获取新数据
			newData := real.GetStockList(c.codes)
			for i := range newData {
				if newData[i]["pct_chg"] != c.data[i]["pct_chg"] {
					c.data[i] = newData[i]
					real.AddSimpleMinute(newData[i])
					// 写入
					err := c.Conn.WriteJSON(newData[i])
					// 错误 关闭连接
					if err != nil {
						Close(c)
					}
				}
			}
		}
	}
}

// Close 关闭连接
func Close(conn *MyConn) {
	fmt.Println("关闭...")
	for i, c := range CListConn {
		if c.uid == conn.uid {
			_ = conn.Conn.Close()
			CListConn = append(CListConn[:i], CListConn[i+1:]...)
			fmt.Println("关闭成功")
			break
		}
	}
}
