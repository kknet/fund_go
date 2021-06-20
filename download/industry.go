package download

import (
	"context"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
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

// Update 更新实时数据至mongo
func Update() {
	info, _ := KlineDB.QueryInterface("select code,sw,industry,area from stock")
	df := dataframe.LoadMaps(info)

	indexes := []string{"code", "name", "pct_chg", "mc", "float_share", "vol", "amount", "main_net", "net"}
	df = df.InnerJoin(AllStock["CN"].Select(indexes), "code")

	for _, item := range df.Maps() {
		err := coll.UpdateId(ctx, item["code"], bson.M{"$set": item})
		if err != nil {
			log.Println("mongo upsert error: ", err)
		}
	}
}

func CalIndustry() {
	for _, idsName := range []string{"sw", "industry", "area"} {
		var results []map[string]interface{}
		err := coll.Aggregate(ctx, mongo.Pipeline{
			// 去掉停牌
			bson.D{{"$match", bson.M{idsName: bson.M{"$ne": "NaN"}, "vol": bson.M{"$gt": 0}}}},
			bson.D{{"$sort", bson.M{"pct_chg": -1}}},
			bson.D{{"$group", bson.M{
				"_id":         "$" + idsName,
				"max_pct":     bson.M{"$first": "$pct_chg"},
				"min_pct":     bson.M{"$last": "$pct_chg"},
				"领涨股":         bson.M{"$first": "$name"},
				"领跌股":         bson.M{"$last": "$name"},
				"count":       bson.M{"$sum": 1},
				"main_net":    bson.M{"$sum": "$main_net"},
				"net":         bson.M{"$sum": "$net"},
				"vol":         bson.M{"$sum": "$vol"},
				"amount":      bson.M{"$sum": "$amount"},
				"float_share": bson.M{"$sum": "$float_share"},
				"mc":          bson.M{"$sum": "$mc"},
				"power":       bson.M{"$sum": bson.M{"$multiply": bson.A{"$mc", "$pct_chg"}}},
			}}},
		}).All(&results)

		if err != nil {
			log.Println(err)
		}
		df := dataframe.LoadMaps(results).Rename("name", "_id")

		// 计算指标
		indexes := []string{"vol", "float_share", "power", "mc"}
		pct := df.Select(indexes).Rapply(func(s series.Series) series.Series {
			value := s.Float()
			return series.Floats([]float64{
				value[0] / value[1] * 10000, // tr
				value[2] / value[3],         // pct_chg
			})
		})
		_ = pct.SetNames("tr", "pct_chg")

		for _, col := range pct.Names() {
			df = df.Mutate(pct.Col(col))
		}
		df = df.Drop([]string{"power", "float_share"})

		Industry[idsName] = df
	}
}
