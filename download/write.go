package download

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"test/myMongo"
	"time"
)

var coll = myMongo.ConnectMongo()

// 更新数据
func writeToMongo(stock []bson.M) {
	start := time.Now()
	err := insertToMongo(stock)

	//var filter []bson.M
	//var update []bson.M
	for _, item := range stock {
		err = coll.UpdateId(ctx, item["code"], bson.M{"$set": item})
		if err != nil {
		}
		//filter = append(filter, bson.M{"_id": item["code"]})
		//update = append(update, bson.M{"$set": item})
	}
	//_, err = coll.UpdateAll(ctx, bson.M{"_id": bson.M{"$in": filter}}, update)
	//if err != nil {
	//}
	fmt.Println(time.Since(start))
}

// 初始化插入
func insertToMongo(stock []bson.M) error {
	for _, item := range stock {
		item["_id"] = item["code"]
		_, err := coll.InsertOne(ctx, item)
		if err != nil {
			return err
		}
	}
	return nil
}
