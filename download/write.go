package download

import (
	"fmt"
	"log"
	"test/myMongo"
	"time"
)

var client = myMongo.ConnectMongo()

// 将下载的数据写入mongo
func writeToMongo(stock []map[string]interface{}, marketType string) error {

	coll := client.Database("stock").Collection(marketType + "Stock")
	err := coll.Drop(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	var docs []interface{}
	for _, item := range stock {
		item["_id"] = item["code"]
		docs = append(docs, item)
	}

	start := time.Now()
	_, err = coll.InsertMany(ctx, docs)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println(marketType, "写入成功,用时：", time.Since(start))
	return nil
}
