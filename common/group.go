package common

import (
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"sync"
)

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
