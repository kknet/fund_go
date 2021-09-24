package common

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

var myClient = &http.Client{Timeout: 5 * time.Second}

// GetAndRead 发送Get请求并读取数据
func GetAndRead(url string) ([]byte, error) {
	res, err := myClient.Get(url)
	if err != nil {
		return nil, errors.New("请求失败！" + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("读取数据失败！" + err.Error())
	}

	return body, nil
}

// GetThsAndRead 同花顺专用
func GetThsAndRead(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)

	//增加header选项
	request.Header.Add("Referer", "http://q.10jqka.com.cn")
	request.Header.Add("Host", "d.10jqka.com.cn")

	res, err := myClient.Do(request)
	if err != nil {
		return nil, errors.New("请求失败！" + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("读取数据失败！" + err.Error())
	}

	return body, nil
}

// GetTuShareAndRead Tushare专用
func GetTuShareAndRead() {

}
