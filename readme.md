---

一个基于 [zapp](https://github.com/zly-app/zapp) 封装的模板库, 提供了常见的组件, log, trace, 提供了默认连接到 apollo
配置等功能.

# 开始

# 配置说明

目前 uapp 主要支持 apollo 配置.

首先程序启动时从 apollo 中 uapp 项目中拉取配置数据. 然后从 apollo 中当前项目中拉取配置数据, 当前项目是指当前运行程序指定的appName. 相当于 uapp 项目中的配置作为一个基础配置数据, 使用者可以在自己的项目中覆盖这些配置数据.

uapp 通过环境变量读取 apollo 配置数据.

| 变量名                     | 是否必须 | 描述                                                                                                   | 默认值    |
| -------------------------- | -------- | ------------------------------------------------------------------------------------------------------ | --------- |
| ApolloAddress              | 是       | apollo-api地址, 多个地址用英文逗号连接                                                                 |           |
| ApolloUAppID               | 否       | uapp 应用名                                                                                            | uapp      |
| ApolloAppId                | 否       | 当前应用名, 应用要覆盖 uapp 的配置                                                                     | \<app名\> |
| ApolloAccessKey            | 否       | 验证key, 优先级高于基础认证                                                                            |           |
| ApolloAuthBasicUser        | 否       | 基础认证用户名, 可用于nginx的基础认证扩展                                                              |           |
| ApolloAuthBasicPassword    | 否       | 基础认证密码                                                                                           |           |
| ApolloCluster              | 否       | 集群名                                                                                                 | default   |
| ApolloAlwaysLoadFromRemote | 否       | 总是从远程获取, 在远程加载失败时不会从备份文件加载                                                     | false     |
| ApolloBackupFile           | 否       | 备份文件名                                                                                             |           |
| ApolloApplicationDataType  | 否       | application命名空间下key的值的数据类型, 支持yaml,yml,toml,json                                         | yaml      |
| ApolloApplicationParseKeys | 否       | application命名空间下哪些key数据会被解析, 无论如何默认的key(frame/components/plugins/services)会被解析 | yaml      |
| ApolloNamespaces           | 否       | 其他自定义命名空间                                                                                     |           |






