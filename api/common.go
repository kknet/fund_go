package apiV1

import (
	"fund_go2/api/real"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// MyConn 连接
type MyConn struct {
	Conn      *websocket.Conn
	codes     []string
	data      []bson.M // 初始化数据 用于判断更新
	uid       string
	CountDown time.Time
	Type      string // 连接类型
}

// ConnList Type: []*MyConn{}
var ConnList = make(map[string][]*MyConn)

// AddToConnList 添加连接
func AddToConnList(ws *websocket.Conn, codes []string, Type string) {
	id, _ := uuid.NewUUID()

	conn := &MyConn{
		Conn: ws, codes: codes, uid: id.String(), Type: Type,
		data: real.GetStockList(codes), CountDown: time.Now(),
	}
	ConnList[Type] = append(ConnList[Type], conn)
}

// Ping 检查连接
func (c *MyConn) Ping() error {
	// 检查连接
	if time.Since(c.CountDown) > time.Minute {
		err := c.Conn.WriteJSON(bson.M{"msg": "ping"})
		// 错误 关闭连接
		if err != nil {
			return err
		} else {
			c.CountDown = time.Now()
			return nil
		}
	}
	return nil
}
