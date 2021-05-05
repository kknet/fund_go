package stock

func GetRank(rankName string) []string {
	// 获取排名前20
	res, err := rdb.ZRange(ctx, "rank:"+rankName, -20, -1).Result()
	if err != nil {
	}
	return res
}
