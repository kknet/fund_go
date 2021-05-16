package download

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"test/myMongo"
	"time"
)

var coll = myMongo.ConnectMongo()

// 更新数据
func writeToMongo(stock []bson.M) {
	start := time.Now()
	err := insertToMongo(stock)

	var filter []bson.M
	var update []bson.M
	for _, item := range stock {
		filter = append(filter, bson.M{"_id": item["code"]})
		update = append(update, bson.M{"$set": item})
	}
	_, err = coll.UpdateAll(ctx, bson.M{"_id": bson.M{"$in": filter}}, update)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(time.Since(start))
}

// 初始化插入
func insertToMongo(stock []bson.M) error {
	var docs []interface{}
	for _, item := range stock {
		item["_id"] = item["code"]
		docs = append(docs, item)
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return err
	}
	return nil
}
