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

// 升级协议
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
	code, ok := c.GetQuery("code")
	if !ok {
		_ = ws.WriteJSON(bson.M{"msg": "必须指定code参数", "status": false})
		return
	}
	codes := strings.Split(code, ",")

	conn := &StockListConn{
		Conn:  newConn(ws),
		codes: codes,
		data:  real.GetStockList(codes),
	}
	StockListConnList = append(StockListConnList, conn)
}

// ConnectItems 详情页连接
func ConnectItems(c *gin.Context) {
	ws, err := upGrade(c)
	if err != nil {
		return
	}
	// 获取参数
	code, ok := c.GetQuery("code")
	if !ok {
		_ = ws.WriteJSON(bson.M{"msg": "必须指定code参数", "status": false})
		return
	}
	// 检查代码合法
	check := real.GetStock(code)
	if len(check) > 0 {
		conn := &StockDetailConn{
			Conn: newConn(ws),
			code: code,
			data: check,
		}
		StockDetailConnList = append(StockDetailConnList, conn)
	} else {
		_ = ws.WriteJSON(bson.M{"msg": "代码不存在", "status": false})
	}
}

// SendCList 推送消息
func SendCList() {
	for _, c := range StockListConnList {
		// 获取新数据
		newData := real.GetStockList(c.codes)
		for i := range newData {
			if newData[i]["pct_chg"] != c.data[i]["pct_chg"] {
				c.data[i] = newData[i]

				// 写入
				err := c.Conn.Conn.WriteJSON(newData[i])
				if err != nil {
					c.Conn.deleteConn()
				}
			}
		}
	}
}

// SendItems 推送详情页
func SendItems() {
	for _, c := range StockDetailConnList {
		// 获取新数据
		newData := real.GetStock(c.code)
		// 有更新
		if newData["vol"] != c.data["vol"] {
			// 详情
			results := real.GetRealTicks(c.code, 50)
			results["items"] = newData
			c.data = newData

			// 写入
			err := c.Conn.Conn.WriteJSON(results)
			if err != nil {
				c.Conn.deleteConn()
			}
		}
	}
}

// ListenChan 监听主函数
func ListenChan() {
	for {
		<-download.MyChan
		SendItems()
		SendCList()
	}
}
