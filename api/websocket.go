package apiV1

import (
	"fund_go2/api/real"
	"fund_go2/download"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 升级协议
func upGrade(c *gin.Context) (*websocket.Conn, error) {
	// 升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(200, gin.H{
			"msg": "协议升级失败", "status": false,
		})
		return nil, err
	}
	return ws, nil
}

// ConnectCList 连接自选表
func ConnectCList(c *gin.Context) {
	ws, err := upGrade(c)
	if err != nil {
		return
	}

	// 获取参数
	codes := strings.Split(c.Query("code"), ",")

	conn := &StockListConn{
		Conn:  newConn(ws),
		codes: codes,
		data:  real.GetStockList(codes),
	}
	listMap[conn.Conn.Id] = conn
}

// ConnectItems 详情页连接
func ConnectItems(c *gin.Context) {
	ws, err := upGrade(c)
	if err != nil {
		return
	}

	code := c.Query("code")
	// 检查代码合法
	check := real.GetStock(code, true)
	if len(check) > 0 {
		conn := &StockDetailConn{
			Conn: newConn(ws),
			code: code,
			data: check,
		}
		// 写入
		err = conn.Conn.Conn.WriteJSON(bson.M{
			"items": check,
		})
		detailMap[conn.Conn.Id] = conn
	} else {
		_ = ws.WriteJSON(bson.M{"msg": "代码不存在", "status": false})
	}
}

// SendCList 推送消息
func (c *StockListConn) SendCList() {
	newData := real.GetStockList(c.codes, true)

	for i := range newData {
		if newData[i]["pct_chg"] != c.data[i]["pct_chg"] {
			// 更新
			err := c.Conn.Conn.WriteJSON(newData[i])
			// 更新错误
			if err != nil {
				delete(listMap, c.Id)
			}
		}
	}
	c.data = newData
}

// SendItems 推送详情页
func (c *StockDetailConn) SendItems() {
	var err error

	newData := real.GetStock(c.code, true)
	// 有更新
	if newData["vol"] != c.data["vol"] {
		c.data = newData

		// 详情信息
		if newData["type"] == "stock" {
			pankou, ticks := real.GetRealTicks(newData)

			err = c.Conn.Conn.WriteJSON(bson.M{
				"items": newData, "ticks": ticks, "pankou": pankou,
			})
		} else {
			err = c.Conn.Conn.WriteJSON(bson.M{
				"items": newData,
			})
		}

		if err != nil {
			delete(detailMap, c.Id)
		}
	}
}

// ListenChan 监听主函数
func ListenChan() {
	for {
		<-download.MyChan
		// 相同code的连接可以同时更新
		for _, c := range detailMap {
			c.SendItems()
		}

		// 对每个连接单独更新
		for _, c := range listMap {
			c.SendCList()
		}
	}
}
