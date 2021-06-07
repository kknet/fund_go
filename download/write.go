package download

import (
	"context"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
)

var realColl = ConnectMongo("AllStock")
var finaColl = ConnectMongo("Fina")

var ctx = context.Background()

// ConnectMongo 连接Mongo
func ConnectMongo(collName string) *qmgo.QmgoClient {
	cli, err := qmgo.Open(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017", Database: "stock", Coll: collName})
	if err != nil {
		panic(err)
	}
	return cli
}

// 更新数据
func writeToMongo(stock []bson.M) {
	var err error
	err = insertToMongo(stock)

	for _, item := range stock {
		err = realColl.UpdateId(ctx, item["code"], bson.M{"$set": item})
		if err != nil {
			//log.Println(err)
		}
	}
}

// 初始化插入
func insertToMongo(stock []bson.M) error {
	for _, item := range stock {
		item["_id"] = item["code"]
		_, err := realColl.InsertOne(ctx, item)
		if err != nil {
			return err
		}
	}
	return nil
}
