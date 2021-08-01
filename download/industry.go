package download

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"sync"
)

var ctx = context.Background()
var RealColl = InitMongo()

// Expression 自定义三元表达式
func Expression(b bool, true interface{}, false interface{}) interface{} {
	if b {
		return true
	} else {
		return false
	}
}

func InitMongo() *qmgo.Collection {
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017"})
	if err != nil {
		panic(err)
	}
	coll := client.Database("stock").Collection("realStock")
	return coll
}

func UpdateMongo(items []map[string]interface{}) {
	group := sync.WaitGroup{}
	group.Add(3)

	myFunc := func(s []map[string]interface{}) {
		myBulk := RealColl.Bulk()

		// 初始化事务
		for i := range s {
			myBulk = myBulk.UpdateId(s[i]["code"], bson.M{"$set": s[i]})
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
	var Id string

	for _, idsName := range []string{"sw", "industry", "area"} {
		// 申万行业的_id为sw_code
		if idsName == "sw" {
			Id = "$sw_code"
			// 其它_id为 idsName + name ("area广东")
		} else {
			Id = "$" + idsName
		}

		err := RealColl.Aggregate(ctx, mongo.Pipeline{
			bson.D{{"$match", bson.M{"marketType": "CN", "type": "stock", "vol": bson.M{"$gt": 0}}}},
			bson.D{{"$sort", bson.M{"pct_chg": -1}}},
			bson.D{{"$group", bson.M{
				"_id":         Id,
				"code":        bson.M{"$first": "$sw_code"},
				"name":        bson.M{"$first": "$" + idsName},
				"max_pct":     bson.M{"$first": "$pct_chg"},
				"领涨股":         bson.M{"$first": "$name"},
				"main_net":    bson.M{"$sum": "$main_net"},
				"net":         bson.M{"$sum": "$net"},
				"vol":         bson.M{"$sum": "$vol"},
				"amount":      bson.M{"$sum": "$amount"},
				"mc":          bson.M{"$sum": "$mc"},
				"fmc":         bson.M{"$sum": "$fmc"},
				"pe_ttm":      bson.M{"$avg": "$pe_ttm"},
				"pb":          bson.M{"$avg": "$pb"},
				"float_share": bson.M{"$sum": "$float_share"},
				"power":       bson.M{"$sum": bson.M{"$multiply": bson.A{"$fmc", "$pct_chg"}}},
			}}},
		}).All(&results)
		if err != nil {
			log.Println("CalIndustry错误: ", err)
		}
		for _, i := range results {
			_, ok := i["_id"].(string)
			if !ok {
				continue
			}
			i["tr"] = float64(i["vol"].(int32)) / i["float_share"].(float64) * 10000
			i["pct_chg"] = i["power"].(float64) / i["fmc"].(float64)
			i["type"] = idsName
			i["marketType"] = "CN"

			if idsName != "sw" {
				delete(i, "code")
			}
			delete(i, "power")
			delete(i, "float_share")

			_, err = RealColl.UpsertId(ctx, i["_id"], i)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
