# fund 金融宝



## 一、网站概述

该网站为个人项目，旨在提供金融行情数据及数据可视化，作为投资参考，不提供任何投资建议

##### 数据来源

[Tushare大数据社区](https://tushare.pro/)、[东方财富网](https://www.eastmoney.com/)

##### 网站地址

https://lucario.ltd



## 二、网站特色

1. 查看实时新闻资讯，包括市场、公司、全球、疫情等等

2. 包含沪深、港、美所有股票详细数据、图表数据，实时数据达到毫秒级更新

   1）详细数据：F10、实时资金

   2）图表数据、分时、K线、资金流向趋势

3. 查看沪深板块、申万行业、地区等实时数据，查看全市场榜单

4. 可登录注册，用户具有积分，积分可以开放部分功能使用权限（暂未上线）



## 三、技术栈

### 1. Django（python）

1. 使用Django restframework框架
2. 行情数据的聚合、计算使用pandas、numpy科学计算库
3. 低频更新数据使用postgres存储，高频实时数据使用MongoDB存储
4. 设置定时脚本每日更新数据库



## 2. Golang

1. 使用gin web框架
2. 使用websocket实时推送最新数据，定义websocket连接列表
3. 实时数据使用dataframe-go、gonum计算，并用MongoDB存储，便于与python联动



### 3. Vue.js

1. 使用vue3.0、vue-router路由插件、vuex状态管理插件等
2. 使用[Element-plus](https://element-plus.gitee.io/#/zh-CN)、[Vant](https://vant-contrib.gitee.io/vant/v3/#/zh-CN)组件库
3. 图表绘制使用[ECharts](https://echarts.apache.org/zh/index.html)
4. 缓存页面，减少加载时间
