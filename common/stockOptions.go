package common

// CListOpt
// StockList Options
type CListOpt struct {
	Codes      []string //代码列表
	MarketType string   // 市场类型
	SortName   string
	Sorted     bool   //排序
	Search     string //搜索
	Size       int    //分页
	Page       int
}

// NewOptByCodes 根据代码列表初始化
func NewOptByCodes(codes []string) *CListOpt {
	return &CListOpt{
		Codes:      codes,
		MarketType: "All", //全市场
	}
}

func NewOptBySearch(search string, marketType string) *CListOpt {
	return &CListOpt{
		Search:     search,
		MarketType: marketType,
	}
}
