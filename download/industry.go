package download

import (
	"context"
	"fund_go2/env"
	_ "github.com/lib/pq"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"sync"
)

var ctx = context.Background()
var realColl *qmgo.Collection

// 初始化MongoDB数据库
func init() {
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: "mongodb://" + env.MongoHost})
	if err != nil {
		panic(err)
	}
	realColl = client.Database("stock").Collection("realStock")
}

// Expression 三元表达式
func Expression(b bool, true interface{}, false interface{}) interface{} {
	if b {
		return true
	}
	return false
}

// 更新实时数据至Mongo
func updateMongo(items []map[string]interface{}) {
	group := sync.WaitGroup{}
	group.Add(3)

	myFunc := func(s []map[string]interface{}) {
		// 使用bulk_write 批量写入
		myBulk := realColl.Bulk()
		for _, i := range s {
			myBulk = myBulk.UpdateId(i["_id"], bson.M{"$set": i})
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

// 聚合计算行业数据
func calIndustry() {
	var res bson.M
	var BKData []bson.M

	for _, idsName := range []string{"industry", "area", "concept"} {
		err := realColl.Find(ctx, bson.M{"type": idsName}).Select(bson.M{"members": 1}).All(&BKData)
		if err != nil {
			log.Println("calIndustry查找"+idsName+"错误: ", err)
		}

		for _, bk := range BKData {
			_ = realColl.Aggregate(ctx, mongo.Pipeline{
				// 不包括新股
				bson.D{{"$match", bson.M{"_id": bson.M{"$in": bk["members"]}, "pct_chg": bson.M{"$lte": 21}}}},
				bson.D{{"$sort", bson.M{"pct_chg": -1}}},
				bson.D{{"$group", bson.M{
					"_id":         nil,
					"max_pct":     bson.M{"$first": "$pct_chg"},
					"领涨股":         bson.M{"$first": "$name"},
					"pct_chg":     bson.M{"$avg": "$pct_chg"},
					"main_net":    bson.M{"$sum": "$main_net"},
					"net":         bson.M{"$sum": "$net"},
					"vol":         bson.M{"$sum": "$vol"},
					"amount":      bson.M{"$sum": "$amount"},
					"mc":          bson.M{"$sum": "$mc"},
					"revenue_yoy": bson.M{"$avg": "$revenue_yoy"},
					"income_yoy":  bson.M{"$avg": "$income_yoy"},
					"roe":         bson.M{"$avg": "$roe"},
					"pe_ttm":      bson.M{"$avg": "$pe_ttm"},
					"pb":          bson.M{"$avg": "$pb"},
					"float_share": bson.M{"$sum": "$float_share"},
				}}},
			}).One(&res)

			res["_id"] = bk["_id"]
			// 更新
			fShare, ok := res["float_share"].(float64)
			if ok {
				res["tr"] = float64(res["vol"].(int32)) / fShare * 10000
				delete(res, "float_share")
			}

			err = realColl.UpdateId(ctx, res["_id"], bson.M{"$set": res})
			if err != nil {
				log.Println("calIndustry更新错误: ", err)
			}
		}
	}
}
