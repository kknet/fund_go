package stock

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
	"test/myMongo"
)

var coll = myMongo.ConnectMongo()

// GetNumbers 获取涨跌分布
func GetNumbers(marketType string) bson.M {

	var temp []bson.M
	_ = coll.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"marketType": marketType, "type": "stock"}}},
		bson.D{{"$group", bson.M{
			"_id":     nil,
			"pct_chg": bson.M{"$push": "$pct_chg"},
			"wb":      bson.M{"$push": "$wb"},
		}}},
	}).All(&temp)

	res := make([]int32, 11)
	pct := temp[0]["pct_chg"].(bson.A)
	wb := temp[0]["wb"].(bson.A)

	for i := range pct {
		p := pct[i].(float64) //涨跌幅pct_chg
		w := wb[i].(float64)  //委比wb

		if p < -7 {
			res[1]++
		} else if p < -5 {
			res[2]++
		} else if p < -3 {
			res[3]++
		} else if p < 0 {
			res[4]++
		} else if p == 0 {
			res[5]++
		} else if p <= 3 {
			res[6]++
		} else if p <= 5 {
			res[7]++
		} else if p <= 7 {
			res[8]++
		} else if p > 7 {
			res[9]++
		}

		if marketType != "CN" {
			if p < -10 {
				res[0]++
			} else if p > 10 {
				res[10]++
			}
		} else {
			if w == -100 {
				res[0]++
			} else if w == 100 {
				res[10]++
			}
		}
	}
	label := []string{"跌停", "<7", "7-5", "5-3", "3-0", "0", "0-3", "3-5", "5-7", ">7", "涨停"}
	if marketType != "CN" {
		label[0] = "<10"
		label[10] = ">10"
	}
	return bson.M{"label": label, "value": res}
}

// GetIndustry 获取板块行情 marketType = CN,HK,US
func GetIndustry(idsName string) []bson.M {
	var results []bson.M
	_ = coll.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"marketType": "CN", "type": "stock", idsName: bson.M{"$nin": bson.A{math.NaN(), nil}}}}},
		bson.D{{"$sort", bson.M{"pct_chg": -1}}},
		bson.D{{"$group", bson.M{
			"_id":     "$" + idsName,
			"max_pct": bson.M{"$first": "$pct_chg"},
			"count":   bson.M{"$sum": 1},
			"领涨股":     bson.M{"$first": "$name"},
			"主力净流入":   bson.M{"$sum": "$main_net"},
			"mc":      bson.M{"$sum": "$mc"},
			"power":   bson.M{"$sum": bson.M{"$multiply": bson.A{"$mc", "$pct_chg"}}},
		}}},
	}).All(&results)
	for _, i := range results {
		i["pct_chg"] = i["power"].(float64) / i["mc"].(float64)
		delete(i, "power")
	}
	return results
}
