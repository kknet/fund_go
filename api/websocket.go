package apiV1

import (
	"fund_go2/api/real"
	"fund_go2/download"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
	"sync"
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
	// 获取参数
	code, ok := c.GetQuery("code")
	if !ok {
		_ = ws.WriteJSON(bson.M{"msg": "必须指定code参数", "status": false})
		return
	}
	codes := strings.Split(code, ",")
	AddToConnList(ws, codes, "clist")
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
	check := real.GetStockList([]string{code})
	if len(check) > 0 {
		AddToConnList(ws, []string{code}, "items")
	} else {
		_ = ws.WriteJSON(bson.M{"msg": "代码不存在", "status": false})
	}
}

// SendCList 推送消息
func SendCList() {
	var err error

	for _, c := range ConnList["clist"] {
		// 获取新数据
		newData := real.GetStockList(c.codes)
		for i := range newData {
			if newData[i]["pct_chg"] != c.data[i]["pct_chg"] {
				c.data[i] = newData[i]
				// 写入
				err = c.Conn.WriteJSON(newData[i])
				// 错误 关闭连接
				if err != nil {
					c.Close()
					break
				}
			}
		}
	}
}

// SendItems 推送详情页
func SendItems() {
	var err error

	for _, c := range ConnList["items"] {
		// 获取新数据
		newData := real.GetStockList(c.codes)[0]
		// 有更新
		if newData["vol"] != c.data[0]["vol"] {
			group := sync.WaitGroup{}
			group.Add(2)
			// 详情
			results := bson.M{"items": newData}
			// 盘口明细
			go func() {
				if newData["marketType"] == "CN" {
					results["pankou"] = real.PanKou(c.codes[0])
				}
				group.Done()
			}()
			// 实时分笔
			go func() {
				results["ticks"], _ = real.GetRealtimeTicks(c.codes[0])
				group.Done()
			}()
			group.Wait()

			c.data[0] = newData
			// 写入
			err = c.Conn.WriteJSON(results)
			if err != nil {
				c.Close()
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
