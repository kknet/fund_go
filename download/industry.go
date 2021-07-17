package download

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"math"
	"sync"
	"xorm.io/xorm"
)

var ctx = context.Background()

var KlineDB = ConnectDB()
var CollDict = InitMongo()

// Expression 自定义三元表达式
func Expression(b bool, true interface{}, false interface{}) interface{} {
	if b {
		return true
	} else {
		return false
	}
}

// ConnectDB 连接数据库
func ConnectDB() *xorm.Engine {
	connStr := "postgres://postgres:123456@127.0.0.1:5432/kline?sslmode=disable"
	db, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

func InitMongo() map[string]*qmgo.Collection {
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017"})
	if err != nil {
		panic(err)
	}
	db := client.Database("stock")

	collDict := map[string]*qmgo.Collection{
		"CN":    db.Collection("CN"),
		"HK":    db.Collection("HK"),
		"US":    db.Collection("US"),
		"Index": db.Collection("Index"), // 存放指数，板块，申万行业
	}
	return collDict
}

func UpdateMongo(items []map[string]interface{}, marketType string) {
	group := sync.WaitGroup{}
	group.Add(3)

	myFunc := func(s []map[string]interface{}) {
		myBulk := CollDict[marketType].Bulk()

		// 初始化事务
		for i := range s {
			myBulk = myBulk.UpdateId(s[i]["code"], bson.M{"$set": s[i]})
			//myBulk = myBulk.UpsertId(s[i]["code"], s[i])
		}
		_, err := myBulk.Run(ctx)
		if err != nil {
			log.Println("更新Mongo错误：", err)
		}
		group.Done()
	}
	length := len(items)
	go myFunc(items[:length/3])
	go myFunc(items[length/3+1 : length/3*2])
	go myFunc(items[length/3*2+1:])
	group.Wait()
}

// CalIndustry 聚合计算板块数据
func CalIndustry() {
	var results []bson.M

	for _, idsName := range []string{"sw", "industry", "area"} {
		err := CollDict["CN"].Aggregate(ctx, mongo.Pipeline{
			// 去掉停牌
			bson.D{{"$match", bson.M{idsName: bson.M{"$nin": bson.A{"NaN", nil, "null", math.NaN()}}, "vol": bson.M{"$gt": 0}}}},
			bson.D{{"$sort", bson.M{"pct_chg": -1}}},
			bson.D{{"$group", bson.M{
				"_id": "$" + idsName,
				//"code":        bson.M{"$first": "$sw_code"},
				"max_pct":     bson.M{"$first": "$pct_chg"},
				"领涨股":         bson.M{"$first": "$name"},
				"main_net":    bson.M{"$sum": "$main_net"},
				"net":         bson.M{"$sum": "$net"},
				"vol":         bson.M{"$sum": "$vol"},
				"amount":      bson.M{"$sum": "$amount"},
				"mc":          bson.M{"$sum": "$mc"},
				"fmc":         bson.M{"$sum": "$fmc"},
				"float_share": bson.M{"$sum": "$float_share"},
				"power":       bson.M{"$sum": bson.M{"$multiply": bson.A{"$mc", "$pct_chg"}}},
			}}},
		}).All(&results)

		if err != nil {
			log.Println("CalIndustry错误: ", err)
		}
		for _, i := range results {
			i["name"] = i["_id"]
			i["_id"] = idsName + i["name"].(string)
			i["type"] = idsName

			i["tr"] = float64(i["vol"].(int32)) / i["float_share"].(float64) * 10000
			i["pct_chg"] = i["power"].(float64) / i["mc"].(float64)

			delete(i, "power")
			delete(i, "float_share")
		}

		for _, i := range results {
			_, err = CollDict["Index"].UpsertId(ctx, i["_id"], i)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
