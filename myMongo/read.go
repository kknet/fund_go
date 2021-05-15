package myMongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

// Read 读取单条数据 返回 bson.M
func Read(cursor *mongo.Cursor) bson.M {
	var result []bson.M
	if err := cursor.All(context.TODO(), &result); err != nil {
		log.Fatal("Read 错误! ", err)
	}
	return result[0]
}

// ReadMany 读取多条数据 返回[]bson.M
func ReadMany(cursor *mongo.Cursor) []bson.M {
	var result []bson.M
	if err := cursor.All(context.TODO(), &result); err != nil {
		log.Fatal("ReadMany 错误! ", err)
	}
	return result
}
