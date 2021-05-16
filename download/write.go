package download

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"test/myMongo"
	"time"
)

var client = myMongo.ConnectMongo()

// 更新数据
func writeToMongo(stock []bson.M) {
	coll := client.Database("stock").Collection("AllStock")

	err := insertToMongo(stock)

	var filter []bson.M
	var update []bson.M
	for _, item := range stock {
		filter = append(filter, bson.M{"_id": item["code"]})
		update = append(update, bson.M{"$set": item})
	}

	start := time.Now()
	_, err = coll.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": filter}}, update)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("写入成功,用时：", time.Since(start))
}

// 初始化插入
func insertToMongo(stock []bson.M) error {
	coll := client.Database("stock").Collection("AllStock")

	var docs []interface{}
	for _, item := range stock {
		item["_id"] = item["code"]
		docs = append(docs, item)
	}
	start := time.Now()
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return err
	}
	fmt.Println("初始化成功,用时：", time.Since(start))
	return nil
}
