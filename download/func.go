package download

import (
	"github.com/go-gota/gota/series"
)

// 生成带初值的Series
func newSeries(value interface{}, name string, len int) series.Series {
	array := make([]interface{}, len)
	for i := range array {
		array[i] = value
	}
	_, ok := value.(string)
	if ok {
		return series.New(array, series.String, name)
	}
	_, ok = value.(float64)
	if ok {
		return series.New(array, series.Float, name)
	}
	_, ok = value.(int)
	if ok {
		return series.New(array, series.Int, name)
	}
	return series.Series{}
}

// Expression 自定义三元表达式
func Expression(b bool, true interface{}, false interface{}) interface{} {
	if b {
		return true
	} else {
		return false
	}
}
