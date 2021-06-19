package download

import (
	"context"
	"github.com/go-gota/gota/dataframe"
	_ "github.com/lib/pq"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"xorm.io/xorm"
)

var ctx = context.Background()

var KlineDB = ConnectDB()
var coll = ConnectMgo().Collection("market")

// ConnectDB 连接数据库
func ConnectDB() *xorm.Engine {
	connStr := "postgres://postgres:123456@127.0.0.1:5432/kline?sslmode=disable"
	db, err := xorm.NewEngine("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}

func ConnectMgo() *qmgo.Database {
	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017"})
	if err != nil {
		panic(err)
	}
	db := client.Database("stock")
	return db
}

// Update 更新实时数据
func Update() {
	info, _ := KlineDB.QueryInterface("select code,sw,industry,area from stock")
	df := dataframe.LoadMaps(info)

	indexes := []string{"code", "name", "pct_chg", "mc", "float_share", "vol", "amount", "main_net", "net"}
	df = df.InnerJoin(AllStock["CN"].Select(indexes), "code")

	for _, item := range df.Maps() {
		_, err := coll.UpsertId(ctx, item["code"], item)
		if err != nil {
			log.Println("mongo upsert error: ", err)
		}
	}
}

func CalIndustry(idsName string) []bson.M {
	var results []bson.M
	_ = coll.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$sort", bson.M{"pct_chg": -1}}},
		bson.D{{"$group", bson.M{
			"_id":         "$" + idsName,
			"max_pct":     bson.M{"$first": "$pct_chg"},
			"领涨股":         bson.M{"$first": "$name"},
			"main_net":    bson.M{"$sum": "$main_net"},
			"net":         bson.M{"$sum": "$net"},
			"vol":         bson.M{"$sum": "$vol"},
			"amount":      bson.M{"$sum": "$amount"},
			"float_share": bson.M{"$sum": "$float_share"},
			"mc":          bson.M{"$sum": "$mc"},
			"power":       bson.M{"$sum": bson.M{"$multiply": bson.A{"$mc", "$pct_chg"}}},
		}}},
	}).All(&results)

	for _, i := range results {
		i["tr"] = i["vol"].(float64) / i["float_share"].(float64) * 10000
		i["pct_chg"] = i["power"].(float64) / i["mc"].(float64)

		delete(i, "power")
		delete(i, "float_share")
	}
	return results
}
