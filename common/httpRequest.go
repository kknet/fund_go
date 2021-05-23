package common

import (
	"io/ioutil"
	"net/http"
)

// myRequest
// 目前只支持Get请求
type myRequest struct {
	url       string
	method    string
	setCookie bool // 配置雪球Cookie
	r         *http.Request
	c         *http.Client
	body      []byte
}

// NewGetRequest
// GET请求
func NewGetRequest(url string, setCookie ...bool) *myRequest {
	res := &myRequest{
		url:       url,
		method:    "GET",
		setCookie: false,
	}
	for _, i := range setCookie {
		if i {
			res.setCookie = true
		}
	}
	return GetHttpRequest(res)
}

func GetHttpRequest(p *myRequest) *myRequest {
	request, err := http.NewRequest(p.method, p.url, nil)
	if err != nil {
		panic(err)
	}
	if p.setCookie {
		request.Header.Add("cookie", "device_id=24700f9f1986800ab4fcc880530dd0ed; s=dk11bk7hr3; cookiesu=301620717341066; remember=1; xq_a_token=986e48f0d816bca49abf998420bd5f7a9df0c506; xqat=986e48f0d816bca49abf998420bd5f7a9df0c506; xq_id_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1aWQiOjM2MTE0MDQxNTUsImlzcyI6InVjIiwiZXhwIjoxNjIzNTA1OTMxLCJjdG0iOjE2MjA5MTM5MzE5OTMsImNpZCI6ImQ5ZDBuNEFadXAifQ.W2LqlQRexNO2VXk0BV91L_uvm9ssWyTYJho51017TI-IRLnkKu6sB35_ZOR1z4XsvnRMSmNlTRDvMKEiapXY4VUu66ySZv3OIzHWaPkxIxBK4cSnL7CFr6CTX0OMAuHuZNHnR-1OJBA5-bPafC47AW0SvJQEs_IBCB83GZK3M859ipuVp_Hn8S0qXbg9v91U-nf4qJXQ4GOT9pjBFQ08u_KagtmfcOfoec23_ejXfrQt_X0F6EKO_w5_LwY0iQmEhE7kM8MiQjOyF6zLOY2JBbnyEkULY4uce5IClP7snpHJp1icydWQsV-eJjlGW9EmVvcDxpIiDvXVG7zfVfjtog; xq_r_token=7193b60d61d2e4db36b3dd1a465837dff68f6400; xq_is_login=1; u=3611404155; bid=3c6bb14598fe9ac45474be34ecb46d45_komyayku; Hm_lvt_1db88642e346389874251b5a1eded6e3=1621264280,1621305028,1621311654,1621324081")
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
