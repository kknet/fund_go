package stock

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"test/myMongo"
	"time"
)

var client = myMongo.ConnectMongo()

// GetNumbers 获取涨跌分布
func GetNumbers(marketType string) []bson.M {
	coll := client.Database("stock").Collection(marketType + "Stock")
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

	opts := options.Aggregate().SetMaxTime(2 * time.Second)

	var results []bson.M
	for i := range label {
		pip := mongo.Pipeline{
			{{"$match", value[i]}},
			{{"$group", bson.M{"_id": label[i], "total": bson.M{"$sum": 1}}}},
		}
		cursor, err := coll.Aggregate(context.TODO(), pip, opts)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, myMongo.Read(cursor))
	}
	return results
}

// GetIndustry 获取板块行情
func GetIndustry(marketType string) []bson.M {
	coll := client.Database("stock").Collection(marketType + "Stock")

	groupStage := bson.D{
		{"$group", bson.M{
			"_id":     "$pct_chg",
			"times":   bson.M{"$sum": 1},
			"总市值":     bson.M{"$sum": "$总市值"},
			"vol":     bson.M{"$sum": "$vol"},
			"amount":  bson.M{"$sum": "$amount"},
			"max_pct": bson.M{"$max": "$pct_chg"},
			"领涨股":     bson.M{"$first": "$name"},
		}},
	}
	opts := options.Aggregate().SetMaxTime(2 * time.Second)
	cursor, err := coll.Aggregate(context.TODO(), mongo.Pipeline{groupStage}, opts)
	if err != nil {
		log.Fatal(err)
	}

	results := myMongo.ReadMany(cursor)
	return results
}
