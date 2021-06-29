package download

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"sync"
	"xorm.io/xorm"
)

var ctx = context.Background()

var KlineDB = ConnectDB()
var realColl = ConnectMgo().Collection("AllStock")

// ConnectDB 连接数据库
func ConnectDB() *xorm.Engine {
	connStr := "postgres://postgres:123456@127.0.0.1:5432/kline?sslmode=disable"
	db, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

func ConnectMgo() *qmgo.Database {
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017"})
	if err != nil {
		panic(err)
	}
	db := client.Database("stock")
	return db
}

func UpdateMongo(items []Stock) {
	group := sync.WaitGroup{}
	group.Add(3)

	myFunc := func(s []Stock) {
		myBulk := realColl.Bulk()
		for i := range s {
			myBulk = myBulk.UpdateId(s[i].Code, bson.M{"$set": s[i]})
		}
		_, _ = myBulk.Run(ctx)
		group.Done()
	}

	length := len(items)
	go myFunc(items[:length/3])
	go myFunc(items[length/3+1 : length/3*2])
	go myFunc(items[length/3*2+1:])
	group.Wait()
}

// UpdateBasic 更新地区 申万行业 板块
func UpdateBasic() {
	info, _ := KlineDB.QueryInterface("select ts_code,industry,sw,sw_code,area from stock")

	for _, i := range info {
		tsCode := i["ts_code"]
		delete(i, "ts_code")
		_ = realColl.UpdateId(ctx, tsCode, bson.M{"$set": i})
	}
}

func CalIndustry() {
	for _, idsName := range []string{"sw", "industry", "area"} {
		var results []bson.M
		err := realColl.Aggregate(ctx, mongo.Pipeline{
			// 去掉停牌
			bson.D{{"$match", bson.M{idsName: bson.M{"$nin": bson.A{"NaN", nil, "null"}}, "vol": bson.M{"$gt": 0}}}},
			bson.D{{"$sort", bson.M{"pct_chg": -1}}},
			bson.D{{"$group", bson.M{
				"_id":         "$" + idsName,
				"code":        bson.M{"$first": "$sw_code"},
				"max_pct":     bson.M{"$first": "$pct_chg"},
				"领涨股":         bson.M{"$first": "$name"},
				"count":       bson.M{"$sum": 1},
				"main_net":    bson.M{"$sum": "$main_net"},
				"net":         bson.M{"$sum": "$net"},
				"vol":         bson.M{"$sum": "$vol"},
				"amount":      bson.M{"$sum": "$amount"},
				"float_share": bson.M{"$sum": "$float_share"},
				"mc":          bson.M{"$sum": "$mc"},
				"fmc":         bson.M{"$sum": "$fmc"},
				"power":       bson.M{"$sum": bson.M{"$multiply": bson.A{"$mc", "$pct_chg"}}},
			}}},
		}).All(&results)

		if err != nil {
			log.Println(err)
		}

		for _, i := range results {
			i["name"] = i["_id"]
			i["tr"] = i["vol"].(float64) / i["float_share"].(float64) * 10000
			i["pct_chg"] = i["power"].(float64) / i["mc"].(float64)

			if idsName != "sw" {
				delete(i, "code")
			}
			delete(i, "_id")
			delete(i, "power")
			delete(i, "float_share")
		}

		Industry[idsName] = results
	}
}
