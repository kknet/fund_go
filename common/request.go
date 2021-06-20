package common

import (
	"io/ioutil"
	"net/http"
)

const (
	COOKIE = "device_id=24700f9f1986800ab4fcc880530dd0ed; s=dk11bk7hr3; cookiesu=301620717341066; remember=1; xq_is_login=1; u=3611404155; bid=3c6bb14598fe9ac45474be34ecb46d45_komyayku; xq_a_token=a93b02bc148d2e262b8a2f630355618f61a85115; xqat=a93b02bc148d2e262b8a2f630355618f61a85115; xq_id_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1aWQiOjM2MTE0MDQxNTUsImlzcyI6InVjIiwiZXhwIjoxNjI0ODc5NTQzLCJjdG0iOjE2MjIyODc1NDM4NDksImNpZCI6ImQ5ZDBuNEFadXAifQ.Qcoj4nqqg5zSEB9GPJn4aQXCokuBRSpDCghhd1Jmr5rKCAJquE7AEIvvI6kMLbcQmmkPNzGvfja8HDKkQDpC4Avr5E9OJuLZSf8VCKjdlJoGAfThiCInUCN6JyIv70kqt68iZB6wC9UcajbkAoxUhu4l5YKxrJf7xVzmsx4UmozEeqxwtPYNB6UInETNuzHxF_STnnQgqawTM6bTtN4oE4U_U4CzJGTc3wCxzfHCQztODTQ-uyCJFet8r1J-flcgHWCpD8TU3Fi9qiABpQfJDFO1YOkBYk1Ygl4IHEKe_0y8-0DekC6BatUMjiFgH7hNqNeLXoYzH9LG5SY65QfF0Q; xq_r_token=53f984ad2accb942b18438b2d7b266073c4567ab; is_overseas=0; Hm_lvt_1db88642e346389874251b5a1eded6e3=1622162076,1622172077,1622287546,1622288784; Hm_lpvt_1db88642e346389874251b5a1eded6e3=1622288784; snbim_minify=true"
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
func NewGetRequest(url string, setCookie ...bool) *myRequest {
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
