package common

// CListOpt 指定代码
type CListOpt struct {
	Codes []string //代码列表
	Chart string   // 简略图表
}

// RankOpt 市场排名
type RankOpt struct {
	MarketType string // 市场类型
	SortName   string
	Sorted     bool //排序
	Page       int64
}
