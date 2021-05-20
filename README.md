部分接口采用go语言重构，实时行情数据用go更新

技术
一、Web
    1. gin框架
    2. websocket长连接
    3. 使用myRequest封装Get请求，链式创建、发起http请求
二、数据
    1. 实时行情数据用MongoDB储存

网站特色
1. 包含沪深、港、美所有股票数据
2. 港美数据毫秒级更新

windows设置代理
1. set GO111MODULE=on
2. set GOPROXY=https://goproxy.cn,direct