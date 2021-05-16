package myMongo

import (
	"context"
	"github.com/qiniu/qmgo"
)

// ConnectMongo 连接Mongo
func ConnectMongo() *qmgo.QmgoClient {
	cli, err := qmgo.Open(context.Background(), &qmgo.Config{Uri: "mongodb://localhost:27017", Database: "stock", Coll: "AllStock"})

	if err != nil {
		panic(err)
	}
	return cli
}
