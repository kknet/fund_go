package stock

func GetRank(rankName string) map[string]float64 {
	// 获取排名前20
	res, err := rdb.ZRangeWithScores(ctx, "rank:"+rankName, -10, -1).Result()
	if err != nil {
	}
	maps := map[string]float64{}
	for i := range res {
		t := res[i]
		maps[t.Member.(string)] = t.Score
	}
	return maps
}
