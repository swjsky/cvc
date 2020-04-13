# 卡券服务上下文环境说明

v 0.0.1

By 欧莱雅IT

## 修订历史

| 版本  | 修订说明       | 提交人   | 生效日期 |
| ----- | -------------- | -------- | -------- |
| 0.0.1 | 初始化创建文档 | Larry Yu |          |

[TOC]

## 引言

### 目的

读者通过阅读本文档，可以了解卡券服务所依赖的运行上下文环境。

本文档的预期读者最好具有软件技术背景。

## 上下文

| 条目               | 说明       | 备注                                                         |
| ------------------ | ---------- | ------------------------------------------------------------ |
| 卡券服务器宿主系统 | CentOS 7.5 |                                                              |
| 卡券服务的数据库   | sqlite     |                                                              |
| oAuth2服务软件     | KeyCloak   | 请参考 https://www.keycloak.org/, <BR>https://hub.docker.com/r/jboss/keycloak/dockerfile |
| OAuth2服务数据库   | mysql      | 8.0.19-1debian9， 请参考 https://github.com/docker-library/mysql/blob/3dfa7a3c038f342b9dec09fa85247bef69ae2349/8.0/Dockerfile |

