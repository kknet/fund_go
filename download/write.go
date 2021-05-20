package download

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
	"test/myMongo"
	"time"
)

var coll = myMongo.ConnectMongo()

// 更新数据
func writeToMongo(stock []bson.M) {
	start := time.Now()
	err := insertToMongo(stock)

	num := 3
	half := len(stock) / num
	// 双协程写入
	group := sync.WaitGroup{}
	group.Add(num)

	tList := [][]bson.M{
		stock[:half], stock[half+1 : half*2+1], stock[half*2+1:],
	}
	for _, i := range tList {
		go func() {
			for _, item := range i {
				err = coll.UpdateId(ctx, item["code"], bson.M{"$set": item})
				if err != nil {
				}
			}
			group.Done()
		}()
	}
	group.Wait()
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
