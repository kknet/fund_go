package common

import (
	"io/ioutil"
	"net/http"
)

// myRequest
// 目前只支持Get请求
type myRequest struct {
	url    string
	method string
	r      *http.Request
	c      *http.Client
	body   []byte
}

// NewGetRequest
// GET请求
func NewGetRequest(url string) *myRequest {
	res := &myRequest{url: url, method: "GET"}
	return GetHttpRequest(res)
}

func GetHttpRequest(p *myRequest) *myRequest {
	request, err := http.NewRequest(p.method, p.url, nil)
	if err != nil {
		panic(err)
	}
	// 保存request到MyRequest中
	p.r = request
	return p
}

// Do *MyRequest的请求方法 返回请求结果
func (r *myRequest) Do() ([]byte, error) {
	if r.c == nil {
		// 初始化Client
		r.c = &http.Client{}
	}
	res, err := r.c.Do(r.r)
	if err != nil {
		return []byte{}, err
	}
	// 关闭连接
	defer res.Body.Close()
	// 读取内容
	body, err := ioutil.ReadAll(res.Body)
	return body, nil
}
