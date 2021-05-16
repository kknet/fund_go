package stock

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"test/myMongo"
)

var coll = myMongo.ConnectMongo()

// GetNumbers 获取涨跌分布 marketType = CN,HK,US
func GetNumbers(marketType string) []bson.M {
	// label 条件搜索
	label := []string{"跌停", "<7", "7-5", "5-3", "3-0", "0", "0-3", "3-5", "5-7", ">7", "涨停"}
	value := []bson.M{
		{"委比": bson.M{"$eq": -100}},
		{"pct_chg": bson.M{"$lt": -7}},
		{"pct_chg": bson.M{"$gte": -7, "$lt": -5}},
		{"pct_chg": bson.M{"$gte": -5, "$lt": -3}},
		{"pct_chg": bson.M{"$gte": -3, "$lt": 0}},
		{"pct_chg": bson.M{"$eq": 0}},
		{"pct_chg": bson.M{"$gt": 0, "$lte": 3}},
		{"pct_chg": bson.M{"$gt": 3, "$lte": 5}},
		{"pct_chg": bson.M{"$gt": 5, "$lte": 7}},
		{"pct_chg": bson.M{"$gt": 7}},
		{"委比": bson.M{"$eq": 100}},
	}
	if marketType != "CN" {
		label[0] = "<10"
		label[10] = ">10"
		value[0] = bson.M{"pct_chg": bson.M{"$lt": -10}}
		value[10] = bson.M{"pct_chg": bson.M{"$gt": 10}}
	}
	var results []bson.M
	var temp []bson.M

	for i := range label {
		matchStage := bson.D{{"$match", value[i]}}
		groupStage := bson.D{
			{"$group", bson.M{"_id": label[i], "total": bson.M{"$sum": 1}}},
		}
		err := coll.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage}).All(&temp)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(temp)
		results = append(results, temp[0])
	}
	return results
}

// GetIndustry 获取板块行情 marketType = CN,HK,US
func GetIndustry(marketType string) []bson.M {
	matchStage := bson.D{{"$match", bson.M{"marketType": marketType}}}
	groupStage := bson.D{
		{"$group", bson.M{
			"_id":     "$pct_chg",
			"总市值":     bson.M{"$sum": "$总市值"},
			"vol":     bson.M{"$sum": "$vol"},
			"amount":  bson.M{"$sum": "$amount"},
			"max_pct": bson.M{"$max": "$pct_chg"},
			"领涨股":     bson.M{"$first": "$name"},
		}},
	}
	var results []bson.M
	err := coll.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage}).All(&results)
	if err != nil {
		log.Fatal(err)
	}
	return results
}
