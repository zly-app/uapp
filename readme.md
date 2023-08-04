<!-- TOC -->

- [开始](#%E5%BC%80%E5%A7%8B)
- [提供的功能](#%E6%8F%90%E4%BE%9B%E7%9A%84%E5%8A%9F%E8%83%BD)
    - [组件](#%E7%BB%84%E4%BB%B6)
    - [插件](#%E6%8F%92%E4%BB%B6)
- [apollo 配置说明](#apollo-%E9%85%8D%E7%BD%AE%E8%AF%B4%E6%98%8E)

<!-- /TOC -->

---

一个基于 [zapp](https://github.com/zly-app/zapp) 封装的模板库, 提供了常见的组件, log, trace, 提供了默认连接到
apollo配置等功能.

# 开始

使用和 `zapp` 没什么区别

```go
app := uapp.NewApp("zapp.test")
defer app.Exit()

c := uapp.GetComponent() // 获取组件
```

# 提供的功能

## 组件

+ [x] [es7](https://github.com/zly-app/component/tree/master/es7)
+ [x] [redis](https://github.com/zly-app/component/tree/master/redis)
+ [x] [sqlx](https://github.com/zly-app/component/tree/master/sqlx)
+ [x] [xorm](https://github.com/zly-app/component/tree/master/xorm)
+ [x] [cache 透明读缓存](https://github.com/zly-app/cache)

## 插件

+ [x] [zipkinotel 链路上报](https://github.com/zly-app/plugin/tree/master/zipkinotel)
+ [x] [honey 日志收集](https://github.com/zly-app/plugin/tree/master/honey)

# apollo 配置说明

目前 `uapp` 主要支持 `apollo` 配置.

首先在 `apollo` 创建一个 `uapp` 项目.

![](src/assets/example/create_uapp.png)

然后在 `application` 中添加好相关配置. 其格式默认为 `yaml`

![](src/assets/example/uapp_config.png)

此时 `uapp` 配置就完成了.

但是现在 `uapp` 还没有接入 `apollo`, 因为没有告诉 `uapp` 如何连接到 `apollo`, 需要在环境变量中配置 `ApolloAddress`,具体环境变量说明如下:

| 变量名                     | 是否必须 | 描述                                                                                                                                      | 默认值    |
| -------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| ApolloAddress              | 是       | apollo-api地址, 多个地址用英文逗号连接                                                                                                    |           |
| ApolloUAppID               | 否       | uapp 应用名                                                                                                                               | uapp      |
| ApolloAppDisable           | 否       | 如果设为true则当前应用不默认获取 apollo 配置, 但是 uapp 不受影响, 也就是 uapp 走apollo配置的同时还可以使用本地配置文件来覆盖 uapp 的配置. | uapp      | false |
| ApolloAppId                | 否       | 当前应用名, 应用要覆盖 uapp 的配置                                                                                                        | \<app名\> |
| ApolloAccessKey            | 否       | 验证key, 优先级高于基础认证                                                                                                               |           |
| ApolloAuthBasicUser        | 否       | 基础认证用户名, 可用于nginx的基础认证扩展                                                                                                 |           |
| ApolloAuthBasicPassword    | 否       | 基础认证密码                                                                                                                              |           |
| ApolloCluster              | 否       | 集群名                                                                                                                                    | default   |
| ApolloAlwaysLoadFromRemote | 否       | 总是从远程获取, 在远程加载失败时不会从备份文件加载                                                                                        | false     |
| ApolloBackupFile           | 否       | 备份文件名                                                                                                                                |           |
| ApolloApplicationDataType  | 否       | application命名空间下key的值的数据类型, 支持yaml,yml,toml,json                                                                            | yaml      |
| ApolloApplicationParseKeys | 否       | application命名空间下哪些key数据会被解析, 无论如何默认的key(frame/components/plugins/services)会被解析                                    |           |
| ApolloNamespaces           | 否       | 其他自定义命名空间                                                                                                                        |           |

环境配置完成后就可以启动程序了.

程序启动时 `upp` 从环境变量中读取 `apollo` 的地址, 如果存在 `${ApolloAddress}`则从 `apollo` 的 `${ApolloUAppID}`项目中拉取配置数据. 然后从 `apollo` 中 `${ApolloAppId}` 项目中拉取配置数据(如果有的话), 如果有重复配置会覆盖掉 `uapp`的配置. 相当于 `uapp` 的配置作为一个基础配置数据.
