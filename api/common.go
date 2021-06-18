package apiV1

import (
	"github.com/gorilla/websocket"
	"time"
)

// MyConn 自选表连接
type MyConn struct {
	Conn      *websocket.Conn
	codes     []string
	data      []map[string]interface{} // 初始化数据 用于判断更新
	uid       string
	CountDown time.Time
}

var CListConn []*MyConn
