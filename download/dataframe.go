package download

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"gonum.org/v1/gonum/mat"
)

type Dataframe struct {
	dataframe.DataFrame
}

// ColIn 判断df中是否存在列
func (df *Dataframe) ColIn(colName string) bool {
	for _, col := range df.Names() {
		if col == colName {
			return true
		}
	}
	return false
}

// RenameDict 根据字典重命名
func (df *Dataframe) RenameDict(maps map[string]string) {
	for key, value := range maps {
		df.Rename(value, key)
	}
}

// Cal 计算
func (df *Dataframe) Cal(s1 series.Series, operation string, s2 series.Series) series.Series {
	v1 := df.calculate(s1, operation, s2)
	return series.New(v1.RawVector().Data, series.Float, "x")
}

// CalAndSet 计算并应用于df
func (df *Dataframe) CalAndSet(s1 series.Series, operation string, s2 series.Series, name string) {
	v1 := df.calculate(s1, operation, s2)
	temp := series.New(v1.RawVector().Data, series.Float, name)
	df.Mutate(temp)
}

// 内部计算操作
func (df *Dataframe) calculate(s1 series.Series, operation string, s2 series.Series) *mat.VecDense {
	v1 := mat.NewVecDense(df.Nrow(), s1.Float())
	v2 := mat.NewVecDense(df.Nrow(), s2.Float())

	switch operation {
	case "+":
		v1.AddVec(v1, v2)
	case "-":
		v1.SubVec(v1, v2)
	case "*":
		v1.MulElemVec(v1, v2)
	case "/":
		v1.DivElemVec(v1, v2)
	}
	return v1
}
