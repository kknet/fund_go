package common

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"go.mongodb.org/mongo-driver/bson"
	"gonum.org/v1/gonum/mat"
	"strings"
	"sync"
)

// RankOpt 市场排名
type RankOpt struct {
	MarketType string // 市场类型
	SortName   string
	Sorted     bool //排序
	Page       int64
}

func GoFunc(data []bson.M, myFunc func(m bson.M)) []bson.M {
	// 多协程
	group := sync.WaitGroup{}
	group.Add(len(data))

	for _, item := range data {
		go func(item bson.M) {
			myFunc(item)
			group.Done()
		}(item)
	}
	group.Wait()
	return data
}

// JoinMapKeys 连接map的key值
func JoinMapKeys(maps map[string]string, concatStr string) string {
	var builder strings.Builder
	for key, _ := range maps {
		builder.WriteString(key)
		builder.WriteString(concatStr)
	}
	str := builder.String()
	str = str[:len(str)-1]
	return str
}

// Expression 自定义三元表达式
func Expression(b bool, true interface{}, false interface{}) interface{} {
	if b {
		return true
	} else {
		return false
	}
}

// InSlice 判断元素在数组中
func InSlice(elem string, arr []string) bool {
	for i := range arr {
		if elem == arr[i] {
			return true
		}
	}
	return false
}

// Operation dataframe列运算
// exp: Operation(df, "total", "a1", "+", "a2") => df['total'] = df['a1'] + df['a2']
func Operation(df dataframe.DataFrame, newCol string, col1 string, operation string, col2 string) dataframe.DataFrame {

	value1 := mat.NewVecDense(df.Nrow(), df.Col(col1).Float())
	value2 := mat.NewVecDense(df.Nrow(), df.Col(col2).Float())

	switch operation {
	case "+":
		value1.AddVec(value1, value2)
	case "-":
		value1.SubVec(value1, value2)
	case "*":
		value1.MulElemVec(value1, value2)
	case "/":
		value1.DivElemVec(value1, value2)
	}

	df = df.Mutate(series.New(value1.RawVector().Data, series.Float, newCol))
	return df
}
