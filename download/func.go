package download

// Expression 自定义三元表达式
func Expression(b bool, true interface{}, false interface{}) interface{} {
	if b {
		return true
	} else {
		return false
	}
}

// SetCol 创建新列
//func (df DataFrame) SetCol(colName string, value interface{}) DataFrame {
//	s, ok := value.(series.Series)
//	if ok {
//		return df.Mutate(series.New(s, s.Type(), colName))
//	}
//
//	array := make([]interface{}, df.Nrow())
//	for i := range array {
//		array[i] = value
//	}
//
//	_, ok = value.(string)
//	if ok {
//		return df.Mutate(series.New(array, series.String, colName))
//	}
//	_, ok = value.(float64)
//	if ok {
//		return df.Mutate(series.New(array, series.Float, colName))
//	}
//	_, ok = value.(int)
//	if ok {
//		return df.Mutate(series.New(array, series.Int, colName))
//	}
//	return df
//}

// RenameDic 重命名
//func (df DataFrame) RenameDic(namesMap map[string]string) DataFrame {
//	for key, value := range namesMap {
//		df = df.Rename(value, key)
//	}
//	return df
//}
