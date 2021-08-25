package apiV1

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Conn 父类Connection
type Conn struct {
	Id         string          // ID
	Conn       *websocket.Conn // 连接实例
	CreateTime time.Time       // 创建时间
}

// StockListConn 自选表连接池
type StockListConn struct {
	*Conn
	codes []string
	data  []bson.M
}

var StockListConnList = make([]*StockListConn, 0)

// StockDetailConn 详情页连接池
type StockDetailConn struct {
	*Conn
	code string
	data bson.M
}

var StockDetailConnList = make([]*StockDetailConn, 0)

// 新建连接
func newConn(ws *websocket.Conn) *Conn {
	id, _ := uuid.NewUUID()
	return &Conn{
		Id:         id.String(),
		Conn:       ws,
		CreateTime: time.Now(),
	}
}

// 删除连接
func (c *Conn) deleteConn() {
	for i := range StockListConnList {
		if StockListConnList[i].Conn.Id == c.Id {
			StockListConnList = append(StockListConnList[:i], StockListConnList[i+1:]...)
			return
		}
	}

	for i := range StockDetailConnList {
		if StockDetailConnList[i].Conn.Id == c.Id {
			StockDetailConnList = append(StockDetailConnList[:i], StockDetailConnList[i+1:]...)
			return
		}
	}
}
