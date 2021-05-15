package myMongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// ConnectMongo 连接Mongo
func ConnectMongo() *mongo.Client {
	// 设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client())
	if err != nil {
		panic(err)
	}
	return client
}
