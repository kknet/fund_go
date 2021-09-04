package common

import (
	"go.mongodb.org/mongo-driver/bson"
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

	for i := range data {
		go func(item bson.M) {
			myFunc(item)
			group.Done()
		}(data[i])
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
