package common

import (
	"go.mongodb.org/mongo-driver/bson"
	"sync"
)

func GoFunc(data []map[string]interface{}, myFunc func(m map[string]interface{})) []map[string]interface{} {
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
