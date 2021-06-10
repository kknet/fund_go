package download

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"log"
)

// 代码格式化
func formatCode(df dataframe.DataFrame) dataframe.DataFrame {
	codes := df.Col("code")
	if &codes == nil {
		log.Println("Col code is not exists!")
		return df
	}
	col := df.Col("marketType")
	if &col == nil {
		log.Println("Col marketType is not exists!")
		return df
	}
	marketType := col.Elem(0).String()
	results := make([]string, codes.Len())

	switch marketType {
	case "CN":
		for i := 0; i < codes.Len(); i++ {
			c := codes.Elem(i).String()
			if c[0] == '0' {
				c += ".SZ"
			} else {
				c += ".SH"
			}
			results[i] = c
		}
	case "CNIndex":
		for i := 0; i < codes.Len(); i++ {
			c := codes.Elem(i).String()
			if c[0] == '0' {
				c += ".SH"
			} else {
				c += ".SZ"
			}
			results[i] = c
		}
	case "HK", "US":
		for i := 0; i < codes.Len(); i++ {
			c := codes.Elem(i).String() + "." + marketType
			results[i] = c
		}
	}
	t := series.Strings(results)
	t.Name = "code"
	df = df.Mutate(t)
	return df
}

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

// Operation 运算操作
// operation = +,-,*,/
// Type(s2) = Series, Int, Float
func Operation(s1 series.Series, operation string, s2 interface{}) series.Series {
	length := s1.Len()
	results := make([]float64, length)

	// Series * Series
	t, ok := s2.(series.Series)
	if ok {
		if t.Len() == length {
			for i := range results {
				results[i] = cal(s1.Elem(i).Float(), operation, t.Elem(i).Float())
			}
			return series.Floats(results)
		} else {
			log.Println("Error: s1.Len() != s2.Len()")
			return series.Series{}
		}
		// Series * Float
	} else {
		for i := range results {

			results[i] = cal(s1.Elem(i).Float(), operation, s2.(float64))
		}
		return series.Floats(results)
	}
}

// 计算
func cal(v1 float64, op string, v2 float64) float64 {
	var res float64
	switch op {
	case "+":
		res = v1 + v2
	case "-":
		res = v1 - v2
	case "*":
		res = v1 * v2
	case "/":
		res = v1 / v2
	}
	return res
}
