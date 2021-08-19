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
	uid   string
	Conn  *websocket.Conn
	codes []string
	data  []bson.M // 初始化数据 用于判断更新
	Type  string
}

// ConnList 所有连接
var ConnList = map[string][]*MyConn{
	"clist": {},
	"items": {},
}

// AddToConnList 添加连接
func AddToConnList(ws *websocket.Conn, codes []string, Type string) {
	id, _ := uuid.NewUUID()
	conn := &MyConn{
		uid:   id.String(),
		Conn:  ws,
		codes: codes,
		Type:  Type,
		data:  real.GetStockList(codes),
	}
	ConnList[Type] = append(ConnList[Type], conn)
	//go conn.Ping()
}

// Ping 检查连接
func (c *MyConn) Ping() {
	last := time.Now()
	for {
		// 每分钟ping
		if time.Since(last) > time.Minute {
			err := c.Conn.WriteJSON(bson.M{"msg": "ping"})
			if err != nil {
				break
			}
			last = time.Now()
		}
		time.Sleep(time.Second * 3)
	}
	defer c.Close()
}

// Close 关闭连接
func (c *MyConn) Close() {
	_ = c.Conn.Close()

	for i, conn := range ConnList[c.Type] {
		if c.uid == conn.uid {
			ConnList[c.Type] = append(ConnList[c.Type][:i], ConnList[c.Type][i+1:]...)
			return
		}
	}
}
