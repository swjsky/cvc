# Larry的工作内容交接

## 卡券服务

注：可以结合intergartion guide for third-parties一起阅读关于卡券的内容。

### 源代码库

https://github.com/iamhubin/loreal.com/tree/master/dit/cmd/coupon-service 目前已经做了转移给 **@iamhubin** 等待approve

### 源代码结构

├── Dockerfile 		制作编译卡券服务的docker 镜像
├── app.go			 app对象，包含卡券服务的一些初始化和各种配置
├── base
│   ├── baseerror.go 			卡券服务错误的结构定义
│   ├── baseerror_test.go
│   ├── config.default.go 		服务的默认配置
│   ├── config.go 					服务配置的数据结构
│   ├── lightutils.go  				轻量的工具函数
│   ├── lightutils_test.go
│   └── requester.go			封装了请求者的信息，比如角色，品牌等
├── config
│   ├── accounts.json		暂未用到，hubin之前的代码遗留
│   └── config.json			程序的运行时配置，将会覆盖默认配置
├── coupon
│   ├── db.coupon.go    							卡券服务的dao层
│   ├── db.coupon_test.go
│   ├── errors.go										卡券服务的各种错误定义
│   ├── logic.coupon.go							卡券结构的一些方法
│   ├── logic.couponservice.go  				卡券服务的一些方法，比如创建卡券，查询卡券，核销卡券
│   ├── logic.couponservice_test.go		 
│   ├── logic.judge.go								规则校验者的定义，如果返回错误，则校验失败。	
│   ├── logic.judge_test.go
│   ├── logic.rulecomposer.go        	 签发卡券时，生成卡券的规则体，比如核销有效时间
│   ├── logic.rulecomposer_test.go
│   ├── message.go                        核销卡券时，可以通知一些相关方核销的信息，这是消息结构。
│   ├── module.coupon.go      卡券以及卡券类型的数据结构
│   ├── module.rule.go     规则以及各个规则细节的结构
│   ├── ruleengine.go     规则引擎，统一调用各个规则体生产规则字符串，以及调用规则校验者校验卡券
│   ├── ruleengine_test.go
│   ├── statics.go     				一些常量
│   ├── test
│   │   ├── coupon_types.json       		定义一些卡券类型，用来api测试的。关于api测试，参考后面章节
│   │   └── rules.json      					用于单元测试的一些规则
│   └── testbase_test.go
├── data
│   ├── data.db            				运行时的sqlite数据库（api测试也会用这个数据库）
│   └── testdata.sqlite				单元测试时的数据库
├── db.go									初始化数据库以及每次启动时执行升级脚本
├── docs
│   ├── authorization\ server\ handbook.md              认证服务器手册
│   ├── context\ of\ coupon\ service.md					部署目标服务器的上下文环境
│   ├── go-live\ handbook.md									上线手册
│   ├── intergartion\ guide\ for\ third-parties.md		第三方开发手册
│   └── technical\ and\ functional\ specification.md   卡券服务功能/技术规格
├── endpoints.debug.go
├── endpoints.gateway.go					暂未涉及，hubin之前的代码遗留
├── endpoints.go						服务的http入口定义，以及做为api成对一些参数进行校验
├── logic.db.go									暂未涉及，hubin之前的代码遗留
├── logic.gateway.go						暂未涉及，hubin之前的代码遗留
├── logic.gateway.upstream.token.go			暂未涉及，hubin之前的代码遗留
├── logic.gateway_test.go
├── logic.go					暂未涉及，hubin之前的代码遗留
├── logic.task.maintenance.go			暂未涉及，hubin之前的代码遗留
├── main.go			主函数入口，载入资源，程序初始化。
├── makefile			
├── message.go			暂未涉及，hubin之前的代码遗留
├── model.brand.go				暂未涉及，hubin之前的代码遗留
├── model.const.go				暂未涉及，hubin之前的代码遗留
├── model.go						暂未涉及，hubin之前的代码遗留
├── module.predefineddata.go			因为卡券类型尚未开发，这里hardcode一些初始化的卡券类型数据。
├── net.config.go				暂未涉及，hubin之前的代码遗留
├── oauth
│   └── oauthcheck.go		校验requester的token
├── pre-defined
│   └── predefined-data.json		hardcode的卡券类型数据，以及规则数据。
├── restful   		该文件夹下的文件暂未涉及，hubin之前的代码遗留
├── sql-migrations
│   └── init-20191213144434.sql			数据库升级脚本
└── task.register.go									暂未涉及，hubin之前的代码遗留

### 重要源代码文件列表

#### endpoints.go

定义了api入口，部分采用了restful风格，方法内会对输入参数进行接口层的校验。

```go
func (a *App) initEndpoints() {
	rt := a.getRuntime("prod")
	a.Endpoints = map[string]EndpointEntry{
		"api/kvstore":            {Handler: a.kvstoreHandler, Middlewares: a.noAuthMiddlewares("api/kvstore")},
		"api/visit":              {Handler: a.pvHandler},
		"error":                  {Handler: a.errorHandler, Middlewares: a.noAuthMiddlewares("error")},
		"debug":                  {Handler: a.debugHandler},
		"maintenance/fe/upgrade": {Handler: a.feUpgradeHandler},
		"api/gw":                 {Handler: a.gatewayHandler},
		"api/events/":            {Handler: longPollingHandler},
		"api/coupontypes":        {Handler: couponTypeHandler},
		"api/coupons/":           {Handler: couponHandler},
		"api/redemptions":        {Handler: redemptionHandler},
		"api/apitester":          {Handler: apitesterHandler},
	}

	postPrepareDB(rt)
}
```

#### db.go

下面的代码段是打包数据库升级脚本文件以及执行升级脚本的代码。

注意：在mac和windows成功执行了从打包文件中读取脚本，但CentOS没有成功，所以目前是手动拷贝的。

```go
migrations := &migrate.PackrMigrationSource{
			Box: packr.New("sql-migrations", "./sql-migrations"),
		}
		n, err := migrate.Exec(env.db, "sqlite3", migrations, migrate.Up)

```

#### sql-migrations/init-20191213144434.sql

这是初始数据库升级脚本。

注意：升级脚本一旦发布，只可增加，不可修改。

#### pre-defined/predefined-data.json

因为卡券类型模块（用户可以通过api创建卡券类型）尚未开发，所以目前是根据业务的需要hard code卡券类型到这里。代码中的第一个卡券类型是测试用。其他6个是正式的卡券。

#### coupon/ruleengine.go

创建卡券时，规则引擎将会检查用户是否可以创建，如果可以，这里会生成各种规则体，附加到卡券上。

核销卡券是，规则引擎检查是否可以核销。

#### coupon/module.rule.go

规则结构，用来描述一个规则，比如核销几次。

#### coupon/logic.rulecomposer.go

rule composer将会根据卡券类型中配置的规则来生成某个规则的规则体，比如

```json
"REDEEM_TIMES": {
    "times": 3
}
```

表示可以核销3次。

注意，卡券结构中的规则体是json格式字符串。

#### coupon/logic.judge.go

judge是每个规则的校验者，如果有问题就返回错误。

比如卡券超兑，会返回 ErrCouponRulesRedeemTimesExceeded

```json
{
	"error-code": 1006,
	"error-message": "coupon redeem times exceeded"
}
```

#### coupon/logic.couponservice.go

相当于传统3层架构的业务层，主要处理卡券相关的业务，签发，查询，核销等。

#### coupon/db.coupon.go

相当于传统3层架构的数据层

#### base/requester.go

表示api请求者身份的。

### 重要的数据结构

#### Rule

rule是描述一个规则，因为尚未开发卡券类型模块，没有对应的数据库表。

这里的结构可以映射为一个数据库表。

其中InternalID是uniqu human readable字符串，比如 REDEEM_TIMES， 表示核销次数规则。

RuleBody是一个json格式的字符串。未来在数据库中应该是一个字符串。

```go
type Rule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	InternalID  string    `json:"internal_id"`
	Description string    `json:"description,omitempty"`
	RuleBody    string    `json:"rule_body"`
	Creator     string    `json:"creator"`
	CreatedTime time.Time `json:"created_time,omitempty" type:"DATETIME" default:"datetime('now','localtime')"`
	UpdatedTime time.Time `json:"updated_time,omitempty" type:"DATETIME"`
	DeletedTime time.Time `json:"deleted_time,omitempty" type:"DATETIME"`
}
```

#### Template

这是一个卡券的原始模板，品牌可以根据末班创建自己的卡券类型。

Creator是创建者，未来可以用于访问控制。

Rules是一个map，保存若干规则，参见 pre-defined/predefined-data.json

```go
type Template struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Creator     string                 `json:"creator"`
	Rules       map[string]interface{} `json:"rules"`
	CreatedTime time.Time              `json:"created_time,omitempty" type:"DATETIME" default:"datetime('now','localtime')"`
	UpdatedTime time.Time              `json:"updated_time,omitempty" type:"DATETIME"`
	DeletedTime time.Time              `json:"deleted_time,omitempty" type:"DATETIME"`
}
```

#### PublishedCouponType

这是根据Template创建的卡券类型。前台系统可以根据卡券类型签发卡券。

TemplateID是基于卡券模板。

Publisher 发布者，未来可以据此进行访问控制。

StrRules 字符串类型的规则。

Rules是 struct类型的规则，用于系统内部处理业务。

```go
type PublishedCouponType struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	TemplateID          string            `json:"template_id"`
	Description         string            `json:"description"`
	InternalDescription string            `json:"internal_description"`
	State               CTState           `json:"state"`
	Publisher           string            `json:"publisher"`
	VisibleStartTime    time.Time         `json:"visible_start_time" type:"DATETIME"`
	VisibleEndTime      time.Time         `json:"visible_end_time"`
	StrRules            map[string]string `json:"rules"`
	Rules               map[string]map[string]interface{}
	CreatedTime         time.Time `json:"created_time" type:"DATETIME" default:"datetime('now','localtime')"`
	DeletedTime         time.Time `json:"deleted_time" type:"DATETIME"`
}
```

#### Coupon

描述一个卡券。

CouponTypeID是PublishedCouponType的ID. 

请参考intergartion guide for third-parties 了解更多

```go
type Coupon struct {
	ID            string
	CouponTypeID  string
	ConsumerID    string
	ConsumerRefID string
	ChannelID     string
	State         State
	Properties    map[string]interface{}
	CreatedTime   *time.Time
}
```

#### Transaction

用户针对卡券做一个操作后，Transaction将描述这一行为。

ActorID是操作者的id。

TransType是操作类型。

ExtraInfo 是操作者附加的信息，用于后期获取后处理前台业务。

```go
type Transaction struct {
	ID          string
	CouponID    string
	ActorID     string
	TransType   TransType
	ExtraInfo   string
	CreatedTime time.Time
}
```

### 重要的接口

#### TemplateJudge

签发卡券时，验证卡券模板。

```go
// TemplateJudge 发卡券时用来验证是否符合rules
type TemplateJudge interface {
	// JudgeTemplate 验证模板
	JudgeTemplate(consumerID string, couponTypeID string, ruleBody map[string]interface{}, pct *PublishedCouponType) error
}
```

#### Judge

核销卡券时，验证卡券。

```go
// Judge 兑换卡券时用来验证是否符合rules
type Judge interface {
	// JudgeCoupon 验证模板
	JudgeCoupon(requester *base.Requester, consumerID string, ruleBody map[string]interface{}, c *Coupon) error
}
```

#### BodyComposer

签发卡券时，生成规则的规则体，用于附加在卡券上。

```go
// BodyComposer 发卡时生成rule的body，用来存在coupon中
type BodyComposer interface {
	Compose(requester *base.Requester, couponTypeID string, ruleBody map[string]interface{}) (map[string]interface{}, error)
}
```

### 单元测试

目前代码中针对coupon文件夹下增加了若干单元测试。

#### 知识准备

阅读单元测试，请先了解：

github.com/smartystreets/goconvey/convey 这是一个可以使用易于描述的方式组织单元测试结构。

bou.ke/monkey 这是一个mock方法的第三方库。

#### 运行单元测试

命令行进入 /coupon， 执行 

```sh
go test   -gcflags=-l
```

因为内联函数的缘故，需要加上 -gcflags=-l

### API测试

目前代码中针对coupon相关的api增加了若干测试。

代码路径： https://github.com/iamhubin/loreal.com/tree/master/dit/cmd/api-tests-for-coupon-service

#### 知识准备

阅读API测试，请先了解：

github.com/smartystreets/goconvey/convey 这是一个可以使用易于描述的方式组织单元测试结构。

github.com/gavv/httpexpect 这是一个api调用并且可以验证结果的第三方库。

#### 执行测试

命令行进入 /api-tests-for-coupon-service， 执行 

```sh
go test
```

### 相关文档

请参阅 https://github.com/iamhubin/loreal.com/tree/master/dit/cmd/coupon-service/docs

### 其他文档

文档压缩包中包含如下几个文件：

| 文件名                                                      | 备注                                    |      |
| ----------------------------------------------------------- | --------------------------------------- | ---- |
| 02_卡券类型申请表单_入会礼_0224.xlsx                        | dennis提交的入会礼卡券                  |      |
| 副本03_客户端申请表单_线上渠道_0219.xlsx                    | dennis提交的客户端                      |      |
| 副本05_用户申请表单_Yiyun_0219.xlsx                         | dennis提交的用户                        |      |
| 卡券类型申请表单_生日礼_0224.xlsx                           | dennis提交的生日礼卡券                  |      |
| 02_Web Application Criticality Determination Form_0219.xlsx | pt测试申请表单                          |      |
| 03_Penetration Test Request Form_0219.xlsx                  | pt测试申请表单                          |      |
| 卡券服务组件架构图.vsdx                                     | 早期文档，用处不大。                    |      |
| 卡券中心最简MVP实施计划.xlsx                                | larry个人做了一点记录，用处不大         |      |
| CardServiceDBSchema.graphml                                 | 数据库设计，用处不大，建议看代码中的DDL |      |
| Loreal卡券服务思维导图.pdf                                  | 早期构思卡券服务功能时的思维导图        |      |
| data flow diagram.pdf                                       | pt测试需要的数据流图                    |      |
| network architecture.pdf                                    | pt测试需要的网络架构图                  |      |

## oAuth2认证服务

认证服务采用了https://hub.docker.com/r/jboss/keycloak。

数据库是https://hub.docker.com/_/mysql。

启动服务的关键命令如下：

```sh
#创建docker的虚拟网络
sudo docker network create keycloak-network 

#启动mysql，注意参数，这不是产线环境参数。
docker run --name mysql -d --net keycloak-network -e MYSQL_DATABASE=keycloak -e MYSQL_USER=keycloak -e MYSQL_PASSWORD=password -e MYSQL_ROOT_PASSWORD=root_password mysql

#启动keycloak，注意参数，这不是产线环境参数。
docker run --name keycloak --net keycloak-network -p 8080:8080 -e KEYCLOAK_USER=yhl10000 -e KEYCLOAK_PASSWORD=Passw0rd jboss/keycloak
```


如何使用认证服务请参阅：卡券服务-相关文档章节。

## 开发测试环境

开发测试环境的服务器从兰伯特那边接过来的。

服务地址：https://gua.e-loreal.cn/#/

服务登录方式请询问hubin。

认证服务请docker ps 相关容器。

卡券服务目录： /home/larryyu/coupon-service。

关于目录结构以及相关功能请咨询hubin。

卡券服务测试服务器是否启动请访问：http://52.130.73.180/ceh/cvc/health

认证服务管理入口：http://52.130.73.180/auth/

## SIT集成测试环境

SIT集成环境用来给供应商开发测试用。

服务登录方式请询问hubin。

服务器有两台，10.162.66.29 和 10.162.66.30 。

卡券服务器目前只用了一台10.162.66.29，类似开发测试环境，包含了认证和卡券两个服务。服务器登录账号目前用的是arvato的账号 **arvatoadmin** 密码是：【请询问hubin】

认证服务请docker ps 相关容器。

卡券服务目录： /home/arvatoadmin/coupon-service。

卡券服务测试服务器是否启动请访问：https://dl-api-uat.lorealchina.com/ceh/cvc/health

认证服务管理入口：跳板机内配置bitvise后，浏览器访问 http://10.162.66.29/auth/

## PRD产线环境

服务登录方式请询问hubin。

服务器有两台：

10.162.65.217 ：认证服务器

10.162.65.218 ：卡券服务器

服务器登录账号目前用的是**appdmin** 密码是：【请询问hubin】

认证服务请docker ps 相关容器。

注意：认证服务数据库在/data1t

卡券服务目录： /home/appadmin/coupon-service。

注意：卡券数据库在/data1t



