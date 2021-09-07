package common

import (
	"errors"
	"io/ioutil"
	"net/http"
)

var myClient = &http.Client{}

// GetAndRead 发送Get请求并读取数据
func GetAndRead(url string) ([]byte, error) {
	// 使用全局client
	res, err := myClient.Get(url)
	defer res.Body.Close()

	if err != nil {
		return nil, errors.New("发送请求失败！" + err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("读取数据失败！" + err.Error())
	}

	return body, nil
}
