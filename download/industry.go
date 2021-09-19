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

// Expression 三元表达式
func Expression(b bool, true interface{}, false interface{}) interface{} {
	if b {
		return true
	}
	return false
}

// InitMongo 初始化连接
func InitMongo() *qmgo.Collection {
	// docker连接: host.docker.internal
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017"})
	if err != nil {
		panic(err)
	}
	coll := client.Database("stock").Collection("realStock")
	return coll
}

// UpdateMongo 更新实时数据至Mongo
func UpdateMongo(items []map[string]interface{}) {
	group := sync.WaitGroup{}
	group.Add(3)

	myFunc := func(s []map[string]interface{}) {
		// 使用bulk_write 批量写入
		myBulk := RealColl.Bulk()
		for _, i := range s {
			myBulk = myBulk.UpdateId(i["code"], bson.M{"$set": i})
		}
		_, err := myBulk.Run(ctx)
		if err != nil {
			log.Println("更新Mongo错误：", err)
		}
		group.Done()
	}
	length := len(items)
	// 三协程
	go myFunc(items[:length/3])
	go myFunc(items[length/3+1 : length/3*2])
	go myFunc(items[length/3*2+1:])
	group.Wait()
}

// CalIndustry 聚合计算行业数据
func CalIndustry() {
	var results []bson.M

	for _, idsName := range []string{"sw", "industry"} {
		err := RealColl.Aggregate(ctx, mongo.Pipeline{
			bson.D{{"$match", bson.M{"marketType": "CN", "type": "stock"}}},
			bson.D{{"$sort", bson.M{"pct_chg": -1}}},
			bson.D{{"$group", bson.M{
				"_id":         "$" + idsName,
				"max_pct":     bson.M{"$first": "$pct_chg"},
				"领涨股":         bson.M{"$first": "$name"},
				"main_net":    bson.M{"$sum": "$main_net"},
				"net":         bson.M{"$sum": "$net"},
				"vol":         bson.M{"$sum": "$vol"},
				"amount":      bson.M{"$sum": "$amount"},
				"mc":          bson.M{"$sum": "$mc"},
				"fmc":         bson.M{"$sum": "$fmc"},
				"revenue_yoy": bson.M{"$avg": "$revenue_yoy"},
				"income_yoy":  bson.M{"$avg": "$income_yoy"},
				"roe":         bson.M{"$avg": "$roe"},
				"pe_ttm":      bson.M{"$avg": "$pe_ttm"},
				"pb":          bson.M{"$avg": "$pb"},
				"float_share": bson.M{"$sum": "$float_share"},
				"power":       bson.M{"$sum": bson.M{"$multiply": bson.A{"$fmc", "$pct_chg"}}},
			}}},
		}).All(&results)
		if err != nil {
			log.Println("CalIndustry聚合错误: ", err)
		}

		for _, i := range results {
			_, ok := i["_id"].(string)
			if !ok {
				continue
			}
			i["tr"] = float64(i["vol"].(int32)) / i["float_share"].(float64) * 10000
			i["pct_chg"] = i["power"].(float64) / i["fmc"].(float64)
			delete(i, "power")
			delete(i, "float_share")

			err = RealColl.UpdateId(ctx, i["_id"], bson.M{"$set": i})
			if err != nil {
				log.Println("CalIndustry更新错误: ", err)
			}
		}
	}
}
