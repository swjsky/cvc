# 卡券服务功能设计说明书

v 0.0.1

By 欧莱雅IT

## 修订历史

| 版本  | 修订说明       | 提交人   | 生效日期 |
| ----- | -------------- | -------- | -------- |
| 0.0.1 | 初始化创建文档 | Larry Yu |          |

[TOC]

## 引言

### 目的

读者通过阅读本文档，可以了解卡券服务的功能以及相关的设计。

本文档的预期读者最好具有软件技术背景。

### 背景

| 条目     | VALUE            | 备注                        |
| -------- | ---------------- | --------------------------- |
| host     | 欧莱雅IT内部开发 |                             |
| 目标用户 | 欧莱雅的其他系统 | 比如campaign tool，云积分等 |
|          |                  |                             |

## 功能需求

卡券服务通过API实现为消费者创建卡券，查询卡券，核销卡券等核心功能。

用户可以查询用户的核销记录。

## 总体设计

### 设计概述

卡券服务使用golang语言开发，采用内嵌类型数据库sqlite保存数据，独立运行，以api的方式提供服务。

卡券服务没有用户界面（UI）。

卡券服务预期部署在linux核心的系统上，比如CentOS。

网络拓扑请参考：network architecture.pdf

数据流请参考：data flow diagram.pdf

### 数据传输

卡券服务以http协议提供访问，作为独立服务，卡券服务本身不提供传输层加密。

卡券服务有且唯一暴露端口（可配置）供客户端连接。

卡券服务部署在内网中，通过反向代理服务（nginx）来暴露服务接口。

nginx会配置https以保障数据在传输层的安全。

### 用户

卡券服务不维护用户，通过解析jwt类型的token来验证用户。

卡券服务验证token的公钥可以用来确保jwt来自信赖的签发来源。

这里可以假设用户为欧莱雅旗下的某些服务或者部门。

关于更多用户相关的信息，请参考 https://www.keycloak.org/

### 消费者

消费者指欧莱雅产品的消费者。

### API

卡券服务以api方式提供服务，尽量使用restful风格设计。

用户通过携带JWT类型token以及消费者的ID调用API为消费者创建，查询，核销卡券。

#### 数据传输格式

数据格式采用json封装。

### 数据表

目前就两张表，请参考下面的DDL:

```sqlite
create table coupons (
	id varchar(36) primary key,
	couponTypeID varchar(36) not null,	-- 卡券类型ID。
	consumerID varchar(256) not null, 	--  唯一确定用户身份。
	consumerRefID varchar(256), 		--  用户身份参考ID,比如腾讯系的openid。
	channelID varchar(256),    			--  创建卡券的渠道。
	state varchar(32) not null, 		-- 卡券当前状态。
	properties varchar(4096), 			-- 保存规则的一些信息，比如可以领取几次。
	createdTime datetime
);

create table couponTransactions (
	id varchar(36) primary key,
	couponID varchar(36) not null,   	--  卡券的ID。
	actorID varchar(256) not null, 		--  操作卡券的用户。
	transType varchar(32) not null, 	-- 业务类型。
	extraInfo varchar(4096),			-- 调用方根据业务需要可以放一些json格式的内容，共后期使用。
	createdTime datetime
);
```



### 卡券类型模块

卡券类型是创建卡券的模板，比如欧莱雅有入会礼和生日礼，那么会有两个卡券类型，分别描述入会礼和生日礼的相关信息，用户可以根据模板为消费者创建卡券。

目前尚未开发卡券类型模块，仅使用hard code方式在代码中提供卡券类型列表。

### 卡券模块

卡券就是一般认为的卡券，比如微信钱包里面的8折促销券。  

卡券模块处理卡券的签发，查询，核销以及相关log查询等功能。

卡券拥有一些基本属性：

| 属性名        | **描述**             | **备注**                                                     |
| ------------- | -------------------- | ------------------------------------------------------------ |
| ID            | 卡券ID               | 每个卡券都有唯一的ID。                                       |
| CouponTypeID  | 卡券类型ID           | 每个卡券属于某个卡券类型，比如入会礼卡券。                   |
| ConsumerID    | 消费者ID             | 卡券中心唯一识别一位消费者的id。比如是CRM内的用户ID.         |
| ConsumerRefID | 消费者的参考ID       | 消费者的参考ID，比如腾讯系的openid。                         |
| ChannelID     | 渠道ID               | 比如某张卡券时由兰蔻线上京东渠道创建的。                     |
| State         | 描述卡券的当前状态。 | 比如一张卡券是否在激活状态。                                 |
| Properties    | 卡券的属性集合       | binding_rule_properties 是和该卡券有关的各种规则限制等。<br />比如一年后过期。 |

#### 尚未解决的问题

​    因为extraInfo是用户提交的信息，可能包含用户手机号码等敏感信息，所以extraInfo需要加密。extraInfo字段尚未使用对称加密加密。

### 错误

一般错误以Http status code来描述，比如400表示提交的数据有错，500表示服务内部有错。

业务错误使用如下格式表示：

```json
{
	"error-code": 2001,
	"error-message": "coupon type not found"
}
```

| **键**        | **值**   | **备注** |
| ------------- | -------- | -------- |
| error-code    | 错误码   |          |
| error-message | 错误消息 |          |

