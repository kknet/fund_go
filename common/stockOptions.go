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

// SearchOpt 搜索
type SearchOpt struct {
	Input string
}

// NewOptByCodes 根据代码列表初始化
func NewOptByCodes(codes []string) *CListOpt {
	return &CListOpt{
		Codes: codes,
	}
}

//func NewOptBySearch(search string, marketType string) *CListOpt {
//	return &CListOpt{
//		Search:     search,
//		MarketType: marketType,
//	}
//}
