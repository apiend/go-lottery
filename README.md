## 抽奖活动服务端 - Golang

### 6 种抽奖小栗子
> 关键点：奖品类型，数量有限，中奖概率，发奖规则  
> 并发安全性问题：互斥锁，队列，CAS递减（atomic）  
> 优化：通过三列减少单个集合的大小

- [年会抽奖](./_demo/1annualMeeting)
- [彩票刮奖](./_demo/2ticket)
- [微信摇一摇](./_demo/3wechatShake)
- [支付宝集福卡](./_demo/4alipayFu)
- [微博抢红包](./_demo/5weiboRedPacket)
- [大转盘](./_demo/6wheel)

### govendor 包管理工具
#### 安装

```shell
go get -u -v github.com/kardianos/govendor
```

#### 初始化

```shell
# 在当前目录生成 vendor 文件夹
govendor init
```

#### 导入项目依赖包

```shell
# 将项目使用的包导入 vendor 文件夹
# 缩写： govendor add +e
govendor add +external
```

#### 本地同步依赖包

```shell
# 重新安装依赖包到 vendor 文件夹
govendor sync
```
    
### 数据设计 - MySQL
#### 奖品表 - gift
| 字段 | 属性 | 说明 |
| --- | --- | --- |
| id | int,pk,auto_increment | 主键 |
| title | varchar(255) | 奖品名称 |
| prize_num | int | 奖品数量：0 无限；>0 限量；<0 无奖品 |
| left_num | int | 剩余奖品数量 |
| prize_code | varchar(50) | 0-9999 标识 100%，0-0 表示万分之一的中奖概率 |
| prize_time | int | 发奖周期：D 天 |
| img | varchar(255) | 奖品图片 |
| display_order | int | 位置序号：小的排在前面 |
| gtype | int | 奖品类型：0 虚拟币；1 虚拟券；2 实物-小奖；3 实物-大奖 |
| gdata | varchar(255) | 扩展数据：如虚拟币数量 |
| time_begin | datetime | 开始时间 |
| time_end | datetime | 结束时间 |
| prize_data | mediumtext | 发奖计划：[[时间 1，数量 1]，[时间 2，数量 2]] |
| prize_begin | datetime | 发奖周期的开始 |
| prize_end | datetime | 发奖周期的结束 |
| sys_status | smallint | 状态：0 正常；1 删除 |
| sys_created | datetime | 创建时间 |
| sys_updated | datetime | 修改时间 |
| sys_ip | varchar(50) | 操作人 IP |

#### 优惠券 - code
| 字段 | 属性 | 说明 |
| --- | --- | --- |
| id | int,pk,auto_increment | 主键 |
| gift_id | int | 奖品 ID，关联 gift 表 |
| code | varcahr(255) | 虚拟券编码 |
| sys_created | datetime | 创建时间 |
| sys_updated | datetime | 更新时间 |
| sys_status | smallint | 状态：0 正常；1 作废；2 已发放 |

#### 抽奖记录表 - result
| 字段 | 属性 | 说明 |
| --- | --- | --- |
| id | int,pk,auto_increment | 主键 |
| gift_id | int | 奖品 ID，关联 gift 表 |
| gift_name | varchar(255) | 奖品名称 |
| gift_type | int | 奖品类型：同 gift.gtype |
| uid | int | 用户 ID |
| username | varchar(50) | 用户名 |
| prize_code | int | 抽奖编号（4 位的随机数） |
| gift_data | varchar(50) | 获奖信息 |
| sys_created | datetime | 创建时间 |
| sys_ip | varchar(50) | 用户抽奖的 IP |
| sys_status | smallint | 状态：0 正常；1 删除；2 作弊 |

#### 用户黑名单表 - black_user
| 字段 | 属性 | 说明 |
| --- | --- | --- |
| id | int,pk,auto_increment | 主键 |
| username | varchar(50) | 用户名 |
| black_time | datetime | 黑名单限制到期时间 |
| real_name | varchar(50) | 联系人 |
| mobile | varchar(50) | 手机号 |
| address | varchar(255) | 联系地址 |
| sys_created | datetime | 创建时间 |
| sys_updated | datetime | 更新时间 |
| sys_ip | varchar(50) | 用户抽奖的 IP |

#### IP 黑名单表 - black_ip
| 字段 | 属性 | 说明 |
| --- | --- | --- |
| id | int,pk,auto_increment | 主键 |
| ip | varchar(50) | IP 地址 |
| black_time | datetime | 黑名单限制到期时间 |
| sys_created | datetime | 创建时间 |
| sys_updated | datetime | 更新时间 |

#### 用户每日次数表 - user_day
| 字段 | 属性 | 说明 |
| --- | --- | --- |
| id | int,pk,auto_increment | 主键 |
| uid | int | 用户 ID |
| day | varchar(8) | 日期：如 20180725 |
| num | int | 次数 |
| sys_created | datetime | 创建时间 |
| sys_updated | datetime | 更新时间 |

### 缓存设计

#### 基本要求
- 目标：提高系统性鞥，减少数据库依赖
- 原则：平衡好"系统性能、开发时间、复杂度"
- 方向：数据读多写少，数据量有限，数据分散

#### 使用位置
- 奖品：数量少，更新频率低，最佳的全量缓存对象
- 优惠券：一次性导入，优惠券编码缓存为 set 类型
- 中奖记录：读写差不多，可以缓存部分统计数据，如：最新中奖记录，最近大奖发放记录等
- 用户黑名单：读多写少，可以按照 uid 散列
- IP 黑名单：类似用户黑名单，可以按照 IP 散列
- 用户每日参与次数：读写次数差异没有用户黑名单那么明显，缓存后的收益不明显

### 系统目录

├── [_demo](./_demo) - 抽奖程序栗子目录  
├── [bootstrap](./bootstrap) - 程序启动相关  
├── [comm](./comm) - 公共代码  
├── [conf](./conf) - 配置相关  
├── [cron](./cron) - 定时任务  
├── [dao](./dao) - 数据库相关操作  
├── [services](./services) - 数据服务类相关  
├── [dataSource](./dataSource) - 数据源  
├── [utils](./utils) - 通用工具  
├── [script](./script) - 独立运行的脚本  
├── [thrift](./thrift) - RPC thrift 相关的  
├── [vendor](./vendor) - 项目依赖包   
└── [weh](./web) - 网站相关  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;├── [controllers](./web/controllers) - 控制器  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;├── [middleware](./web/middleware) - 中间件  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;├── [public](./web/public) - 静态文件  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;├── [routes](./web/routes) - 路由  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;├── [views](./web/views) - 模板  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;└── [viewModels](./web/viewModels) - 模板数据模型  

### commit 规范
- [dev] + 新增的内容
- [upd] + 更新的内容

未完待续。。。

如有错误，务必指出，谢谢 ^.^ !


# lottery

### 包依赖

~~~
go get github.com/kataras/iris
go get github.com/kataras/iris/mvc
go get github.com/iris-contrib/middleware/cors
go get github.com/kataras/iris/httptest
go get github.com/go-xorm/xorm
go get github.com/go-sql-driver/mysql
go get github.com/kataras/iris/middleware/recover
go get github.com/kataras/iris/middleware/logger
~~~

### 翻墙配置

~~~
replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.34.0
	github.com/go-tomb/tomb => gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	go.opencensus.io => github.com/census-instrumentation/opencensus-go v0.19.0
	go.uber.org/atomic => github.com/uber-go/atomic v1.3.2
	go.uber.org/multierr => github.com/uber-go/multierr v1.1.0
	go.uber.org/zap => github.com/uber-go/zap v1.9.1

	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20181001203147-e3636079e1a4
	golang.org/x/lint => github.com/golang/lint v0.0.0-20181026193005-c67002cb31c3
	golang.org/x/net => github.com/golang/net v0.0.0-20180826012351-8a410e7b638d
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync => github.com/golang/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20181116152217-5ac8a444bdc5
	golang.org/x/text => github.com/golang/text v0.3.0
	golang.org/x/time => github.com/golang/time v0.0.0-20180412165947-fbb02b2291d2
	golang.org/x/tools => github.com/golang/tools v0.0.0-20181219222714-6e267b5cc78e
	google.golang.org/api => github.com/googleapis/google-api-go-client v0.0.0-20181220000619-583d854617af
	google.golang.org/appengine => github.com/golang/appengine v1.3.0
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20181219182458-5a97ab628bfb
	google.golang.org/grpc => github.com/grpc/grpc-go v1.17.0
	gopkg.in/alecthomas/kingpin.v2 => github.com/alecthomas/kingpin v2.2.6+incompatible
	gopkg.in/mgo.v2 => github.com/go-mgo/mgo v0.0.0-20180705113604-9856a29383ce
	gopkg.in/vmihailenco/msgpack.v2 => github.com/vmihailenco/msgpack v2.9.1+incompatible
	gopkg.in/yaml.v2 => github.com/go-yaml/yaml v0.0.0-20181115110504-51d6538a90f8
	labix.org/v2/mgo => github.com/go-mgo/mgo v0.0.0-20160801194620-b6121c6199b7
	launchpad.net/gocheck => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
)
~~~
