package download

import (
	"github.com/go-gota/gota/series"
	"log"
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

// Operation 运算操作
// operation = +, -, *, /
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

// OperationString 字符串运算操作
// Example: x=1, y=a+b, x=a*b (暂不支持加括号的运算)
//func OperationString(str string, df dataframe.DataFrame) series.Series {
//	//去除所有空格
//	str = strings.Replace(str," ","",-1)
//	equ := strings.Split(str, "=")
//
//	//匹配
//	var left, right interface{}
//	var op string
//	var err error
//
//	for _,op = range []string{"+", "-", "*", "/"} {
//		items := strings.Split(equ[1], op)
//		if len(items) >= 2 {
//			left, err = strconv.ParseFloat(items[0], 32)
//			if err != nil {
//				left = df.Col(items[0])
//			}
//
//			right, err = strconv.ParseFloat(items[1], 32)
//			if err != nil {
//				right = df.Col(items[1])
//			}
//			break
//		}
//	}
//	//计算
//	length := df.Nrow()
//	_, l1 := left.(series.Series)
//	_, l2 := left.(float64)
//	_, r1 := right.(series.Series)
//	_, r2 := right.(float64)
//
//
//}

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
