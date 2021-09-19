package common

import (
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"sync"
	"unsafe"
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

// Str2bytes 使用unsafe包转换string与byte
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
