package apiV1

import (
	"errors"
	"io/ioutil"
	"net/http"
)

type MyRequest struct {
	http.Request
	conn *http.Client
}

// NewRequest 实例化
func NewRequest() *MyRequest {
	return &MyRequest{conn: &http.Client{}}
}

// DoAndRead 发送请求并读取数据（需要初始化实例）
func (r *MyRequest) DoAndRead(url string) ([]byte, error) {
	res, err := r.conn.Get(url)
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, errors.New("发送请求失败！" + err.Error())
	} else {
		return body, nil
	}
}

// GetAndRead 发送请求并读取数据
func GetAndRead(url string) ([]byte, error) {
	res, err := http.Get(url)
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, errors.New("发送请求失败！" + err.Error())
	} else {
		return body, nil
	}
}
