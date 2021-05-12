package download

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// mongo
var client = connectToMongo()

// 连接mongoDB数据库
func connectToMongo() *mongo.Client {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	return client
}

// 写入mongoDB
func writeToMongo(stocks []map[string]interface{}, marketType string) {
	begin := time.Now()
	collection := client.Database("stock").Collection(marketType)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var docs []interface{}
	for _, item := range stocks {
		item["_id"] = item["code"]
		docs = append(docs, item)
	}
	//删除collection
	err := collection.Drop(ctx)
	if err != nil {
		log.Println(err)
	}
	//插入所有数据
	_, err = collection.InsertMany(ctx, docs)
	if err != nil {
		log.Println(err)
	}
	sl := time.Since(begin)
	fmt.Println(sl)
}
