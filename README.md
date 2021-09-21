# fund 金融宝

## 一、网站概述

该网站为个人项目，旨在提供金融行情数据及数据可视化，为用户提供直观的实时、历史数据展示，可作为投资参考，不构成任何投资建议

##### 数据来源

[Tushare大数据社区](https://tushare.pro/)、[东方财富网](https://www.eastmoney.com/)、[AKShare](https://www.akshare.xyz/)、[同花顺财经](http://www.10jqka.com.cn/)

##### 网站地址

https://lucario.ltd



## 二、网站特色

- 查看实时新闻资讯，包括市场、公司、全球、疫情等等

- 可视化查看个股行情，包含沪深、港、美所有股票数据，实时数据毫秒级更新

- 查看全市场榜单，以及沪深板块、行业、地区数据

- 根据财务、行情指标选股

- 注册、登录成为用户，用户具有积分



## 三、网站技术

项目环境：Gin(Golang) + Django(Python) + Vue + MongoDB + Redis + Postgresql

### Golang

1. 使用gin web框架，自定义中间件拦截一些非法请求，复用功能
2. 用户模块使用xorm + validator 验证表单、操作数据库，使用jwt技术 + Redis验证token
3. 使用Redis记录股票热度（访问次数）
4. 实时数据使用dataframe-go、gonum计算，使用MongoDB存储，并使用websocket实时推送
5. 使用docker-compose打包gin+Redis+Postgresql，一键部署



### Django（python）

1. 使用Django restframework框架
2. 行情数据的聚合、计算使用pandas、numpy科学计算库
3. 历史数据使Postgresql存储，高频更新数据使用MongoDB存储
4. 设置定时脚本盘前、盘后更新数据库

### Vue.js

1. 使用vue3.0、vue-router路由插件、vuex状态管理插件等
2. 使用[Element-plus](https://element-plus.gitee.io/#/zh-CN)、[Vant（移动端）](https://vant-contrib.gitee.io/vant/v3/#/zh-CN)组件库
3. 图表绘制使用[ECharts](https://echarts.apache.org/zh/index.html)
4. 缓存页面，减少加载时间
