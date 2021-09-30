package apiV1

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

// StockDetail Data Connection
var detailMap = map[string]*StockDetailConn{}

// StockList Data Connection
var listMap = map[string]*StockListConn{}

type Conn struct {
	Id   string          // ID
	Conn *websocket.Conn // 连接实例
}

// StockListConn 自选表连接池
type StockListConn struct {
	*Conn
	codes []string
	data  []bson.M
}

// StockDetailConn 详情信息连接
type StockDetailConn struct {
	*Conn
	code string
	data bson.M
}

// 新建连接
func newConn(ws *websocket.Conn) *Conn {
	id, _ := uuid.NewUUID()
	return &Conn{
		Id:   id.String(),
		Conn: ws,
	}
}
