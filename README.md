部分接口采用go语言重构，实时行情数据用go更新

网站地址
https://lucario.ltd

技术
1. 使用gin web框架
2. 使用websocket实时推送最新数据
3. 使用myRequest封装Get请求，链式创建、发起http请求
4. 使用struct替代map[string]interface{}解析json数据, 大大提高解析速度

数据
1. 实时行情数据使用dataframe-go储存

网站特色
1. 包含沪深、港、美所有股票数据
2. 港美数据毫秒级更新
3. 查看沪深板块、申万行业、地区等实时数据
4. 可登录注册，用户具有积分，积分可以开放部分功能使用权限(未上线)