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
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(200, gin.H{
			"msg": "连接失败", "status": false,
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
	defer ws.Close()

	// 获取参数
	code, ok := c.GetQuery("code")
	if !ok {
		_ = ws.WriteJSON(bson.M{"msg": "必须指定code参数", "status": false})
		return
	}

	codes := strings.Split(code, ",")
	oldData := real.GetStockList(codes)
	for {
		<-download.MyChan
		newData := real.GetStockList(codes)
		for i := range newData {
			if newData[i]["pct_chg"] != oldData[i]["pct_chg"] {
				oldData[i] = newData[i]
				// 写入
				err = ws.WriteJSON(newData[i])
				if err != nil {
					return
				}
			}
		}
	}
}

// ConnectItems 详情页连接
func ConnectItems(c *gin.Context) {
	ws, err := upGrade(c)
	if err != nil {
		return
	}
	defer ws.Close()

	// 获取参数
	code, ok := c.GetQuery("code")
	if !ok {
		_ = ws.WriteJSON(bson.M{"msg": "必须指定code参数", "status": false})
		return
	}

	// 检查代码是否存在
	check := real.GetStockList([]string{code})
	if len(check) <= 0 {
		return
	}

	oldData := check[0]
	for {
		<-download.MyChan
		newData := real.GetStockList([]string{code})[0]
		// 有更新
		if newData["vol"].(int32) > oldData["vol"].(int32) {
			// 详情
			results := real.GetRealTicks(code, 50)
			results["items"] = newData
			// 写入
			err = ws.WriteJSON(results)
			if err != nil {
				return
			}
		}
	}
}
