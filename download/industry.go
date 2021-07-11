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
var CollDict = InitMongo()

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

func UpdateMongo(items []Stock, marketType string) {
	group := sync.WaitGroup{}
	group.Add(3)

	myFunc := func(s []Stock) {
		myBulk := CollDict[marketType].Bulk()

		// 初始化事务
		for i := range s {
			// 合法性
			if marketType != "Index" && s[i].TotalShare <= 0 {
				continue
			}
			myBulk = myBulk.UpdateId(s[i].Code, bson.M{"$set": s[i]})
			//myBulk = myBulk.UpsertId(s[i].Code, s[i])
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

// UpdateBasic 更新地区 申万行业 板块
func UpdateBasic() {
	info, _ := KlineDB.QueryInterface("select ts_code,industry,sw,sw_code,area from stock")

	for _, i := range info {
		tsCode := i["ts_code"]
		delete(i, "ts_code")
		_ = CollDict["CN"].UpdateId(ctx, tsCode, bson.M{"$set": i})
	}
}

// CalIndustry 聚合计算板块数据
func CalIndustry() {
	var results []bson.M

	for _, idsName := range []string{"sw", "industry", "area"} {
		err := CollDict["CN"].Aggregate(ctx, mongo.Pipeline{
			// 去掉停牌
			bson.D{{"$match", bson.M{idsName: bson.M{"$nin": bson.A{"NaN", nil, "null"}}, "vol": bson.M{"$gt": 0}}}},
			bson.D{{"$sort", bson.M{"pct_chg": -1}}},
			bson.D{{"$group", bson.M{
				"_id":         "$" + idsName,
				"code":        bson.M{"$first": "$sw_code"},
				"max_pct":     bson.M{"$first": "$pct_chg"},
				"领涨股":         bson.M{"$first": "$name"},
				"count":       bson.M{"$sum": 1},
				"main_in":     bson.M{"$sum": "$main_in"},
				"main_out":    bson.M{"$sum": "$main_out"},
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
			i["type"] = idsName
			i["tr"] = i["vol"].(float64) / i["float_share"].(float64) * 10000
			i["pct_chg"] = i["power"].(float64) / i["mc"].(float64)
			i["main_net"] = i["main_in"].(float64) + i["main_out"].(float64)
			i["main_pct"] = i["main_net"].(float64) / (i["main_in"].(float64) - i["main_out"].(float64)) * 100

			delete(i, "_id")
			delete(i, "power")
			delete(i, "float_share")
		}

		for _, i := range results {
			_, err = CollDict["Index"].UpsertId(ctx, idsName+i["name"].(string), i)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
