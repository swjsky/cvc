# 用户认证服务上线/运营手册

v 0.0.1

by 欧莱雅IT

## 修订历史

| 版本  | 修订说明       | 提交人   | 生效日期 |
| ----- | -------------- | -------- | -------- |
| 0.0.1 | 初始化创建文档 | Larry Yu |          |
|       |                |          |          |

[TOC]



## 引言

### 目前的部署情况

SIT 环境：https://dl-api-uat.lorealchina.com/auth/realms/Lorealcn/protocol/openid-connect/token
PRD 环境：https://dl-api.lorealchina.com/auth/realms/Lorealcn/protocol/openid-connect/token

### 目的

用户认证服务将会提供欧莱雅内部用户的账号维护，以及为依赖用户认证服务的应用签发令牌和核验令牌。

为方便相关人员理解系统以及如何操作，本文档介绍如何上线部署用户认证服务，以及后期运营。

### 准备工作

如果阅读者是部署人员，需要了解docker，mysql，linux环境。

如果是配置人员，需要了解oAuth2。

## 部署

请参考 go-live handbook

【注意】部署时，注意初始化管理员账号。

部署成功后，可以打开{host}/auth/ 来测试是否部署成功。

## 导入realm[可选]

因为在测试环境已经创建了realm，为了简化操作，可以直接导入已经存在的realm。

## 配置realm

如果没有导入一个现有的realm，那么需要创建一个新的。

### Login 标签页，

- 配置用email登录
- 外部请求使用SSL
- 其他可以关闭

### Keys标签页

查看RS256的公约，用来给卡券服务作为认证之用。

### Tokens标签页

一般默认就可，除非特别配置。

### 导出配置

为了方便管理以及迁移数据，管理员可以有限导出realm的数据。包括：

| 项目     | 描述                                                         |      |
| -------- | ------------------------------------------------------------ | ---- |
| 组和角色 | 根据业务的需要，可以创建一些组，比如具有相同角色的人可以放在一个组里面。<BR>角色用来描述一个用户可以做什么事情。 |      |
| 客户端   | 客户端是用来描述一个接入oAuth2服务的程序或者服务。比如欧莱雅内部campaign tool。<BR>客户端功能可以有效区分不同的应用，分别配置访问资源的权限，更大可能保护用户的资源等。 |      |
|          |                                                              |      |

注意：**管理员无法导出用户信息。**

## 管理Clients

这里的客户端，也就是应用，目前我们有campaign tool和云积分，也就是说，至少有两个客户端需要配置。

原则上，为了安全，一个单独的服务，就是一个client。

### 配置Tab

- 配置Client ID,将会交付给应用开发商。
- Enabled标志应用是否被启用。
- Consent required 选择False. 【注意】因为用户是欧莱雅内部员工，所以此处不需要Consent，这不是常用的选择。
- Client Protocol 选择 openid-connect。
- Access Type 选择 confidential。
- Authorization Enable 选择false。

### Crendentias Tab

Client Authenticator  选择 client id and secret, 然后生成一个secret。【**注意**】**<u>这个secret是保密内容，请使用安全的方式交付给应用开发商</u>**。

### Mappers Tab

这里为一个client配置一个mapper，将会在用户的token里增加一些项，比如用户所属的品牌信息。

新建一个mapper，取一个有意义的名字，比如**用户所属品牌**，

打开mapper，编辑各个属性：

-  Protocol ：选择openid-connect
-  Mapper Type ：选择 User Attribute
-  User Attribute  ：**brand**，【 注意】这是预先定义好的，不要修改成其他的。
-  Token Claim Name ：**brand**，【 注意】这是预先定义好的，不要修改成其他的。
-  Claim JSON Type ：String
-  Add to ID token  ：选择ON
-  Add to access token : 选择ON
-  Add to userinfo  : 选择ON

## 角色

新建三个角色，如下表：

| 角色名          | 用途                                   | 备注                                                         |
| --------------- | -------------------------------------- | ------------------------------------------------------------ |
| coupon_issuer   | 可以签发卡券                           | 比如campaign tool需要发券，那么内置的用户需要这个角色        |
| coupon_listener | 可以监听卡券服务的事件，比如核销事件。 | 比如campaign tool想得知哪个消费者核销了哪个券，可以配置这个角色，然后长轮询核销卡券的事件。 |
| coupon_redeemer | 可以核销卡券                           | 比如云积分需要发券，那么内置的用户需要这个角色               |

## 组（Groups）

因为有一些业务需求是不允许夸品牌兑换，通过配置组可以实现用户品牌的区分。

用户认证服务里面的组相当于欧莱雅的品牌。

比如新建组：**兰蔻** 。然后在**属性Tab**里面增加一个属性：

| Key   | Value   | 备注                                                         |
| :---- | :------ | :----------------------------------------------------------- |
| brand | LANCOME | key 必须是brand，卡券中心将依赖这个配置。Value配置成有意义的值，一旦配置好后，因为业务依赖，将很难被更改掉。 |

## 用户

### 新增用户

目前除了Username是必选项，其他默认也行，但为了管理，最好丰富下其他信息。

### 配置用户

#### Role Mappings

这里配置用户的角色，根据业务可选前面提到的**角色**。将角色添加到**Assigned Roles**.

#### Groups

给用户配置组别，前面提到有些服务需要组别来判断用户的品牌属性。

在右侧可选的组别中根据业务选择一个组，【注意】只选择一个组。

## 日志管理

在keycloak的管理-事件模块，可以管理日志。

### 开启日志

在**Config** tab，可以分别开启登录和管理两类事件日志。

### 查看日志

在**登录事件**和**管理时间**两个tab，可以看到两类事件的详情。













































