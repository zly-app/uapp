# Config — 配置文件与配置变更监听

> 导入: `github.com/zly-app/zapp/config`

## 功能概览

- 多来源配置加载: 命令行 > WithViper > WithConfig > WithFiles > WithApollo > 默认文件
- 按名解析配置到结构体
- 配置热加载（配合 Apollo）
- 支持 Apollo 命名空间

## 配置结构

zapp 配置分为以下部分：

```yaml
frame:              # 框架配置 (debug/env/name/instance/log等)
components:         # 组件配置 (redis/sqlx/grpc/cache/gpool等)
  redis:
    default: { ... }
  sqlx:
    default: { ... }
services:           # 服务配置 (grpc/cron等)
  grpc: { ... }
plugins:            # 插件配置 (otlp/prometheus等)
filters:            # 过滤器配置
  config:
    base: { ... }
```

## 使用方式

### 1. 解析业务配置

```go
import "github.com/zly-app/zapp/config"

const ConfigKey = "my-service"  // 配置 key

type Config struct {
    RedisName string
    SqlxName  string
    MaxRetry  int
}

var Conf = &Config{}

func Init() error {
    err := config.Conf.Parse(ConfigKey, Conf, true)
    if err != nil {
        return err
    }
    Conf.Check()  // 校验+填充默认值
    return nil
}

func (c *Config) Check() {
    if c.RedisName == "" { c.RedisName = "default" }
    if c.SqlxName == "" { c.SqlxName = "default" }
    if c.MaxRetry == 0 { c.MaxRetry = 3 }
}
```

### 2. 解析组件/插件/服务配置

```go
// 解析指定组件配置（第一个参数为 ComponentType，非字符串）
var redisCfg redis.RedisConfig
err := config.Conf.ParseComponentConfig(redis.DefaultComponentType, "default", &redisCfg, true)

// 解析插件配置
var otlpCfg otlp.Config
err := config.Conf.ParsePluginConfig(otlp.DefaultPluginType, &otlpCfg, true)

// 解析服务配置
var grpcCfg grpc.ServiceConfig
err := config.Conf.ParseServiceConfig(grpc.DefaultServiceType, &grpcCfg, true)
```

### 3. 配置变更监听

配置变更监听基于 Provider 机制（如 Apollo 插件自动注册为 Provider），通过 `WatchKey` 系列 API 实现：

```go
import "github.com/zly-app/zapp/config"

// 方式1: 原始字节监听（适合简单值）
watcher := config.WatchKey("my-namespace", "someKey")
watcher.AddCallback(func(isInit bool, oldData, newData []byte) {
    log.Info("配置变更", zap.Bool("isInit", isInit), zap.String("new", string(newData)))
})
// 读取当前值
strVal := watcher.GetString()
intVal := watcher.GetInt(0)  // 可带默认值

// 方式2: 结构化监听 — 泛型，自动反序列化（默认 JSON）
watcher := config.WatchKeyStruct[*MyConfig]("my-namespace", "my-key")
watcher.AddCallback(func(isInit bool, oldData, newData *MyConfig) {
    // newData 已自动反序列化
    log.Info("配置变更", zap.Any("new", newData))
})
// 读取当前结构化值
cfg := watcher.Get()

// 方式3: 指定格式的结构化监听
jsonWatcher := config.WatchJson[*MyConfig]("ns", "key")  // JSON 格式
yamlWatcher := config.WatchYaml[*MyConfig]("ns", "key")  // YAML 格式
```

Watch 选项：

```go
// 指定 Provider（默认使用 default provider，即 Apollo）
config.WithWatchProvider("custom-provider")

// 指定结构化类型（WatchKeyStruct 默认为 JSON）
config.WithWatchStructYaml()
config.WithWatchStructJson()
```

### 4. 注册 Apollo 命名空间

Apollo 的 `application` 命名空间下，默认只解析以下 key：`frame`、`components`、`plugins`、`filters`、`services`。
如果你的业务配置放在同一个 `application` 命名空间下的自定义 key 中（而非独立命名空间），需要在 `NewApp` 之前注册，否则该 key 的配置不会被解析：

```go
// 在 main.go 中，NewApp 之前
// 假设 Apollo application 命名空间下有一个 key 为 "my-service" 的配置
config.RegistryApolloNeedParseNamespace("my-service")

// 如果业务配置放在 Apollo 的独立命名空间（非 application）中，则不需要调用此函数
// Apollo 的 Namespaces 配置会自动加载整个命名空间
```

### 5. 获取 Viper 实例

```go
vi := config.Conf.GetViper()
```

## 默认配置文件

zapp 默认从以下位置加载配置文件：

- `./configs/default.yaml`
- `./configs/default.yml`
- `./configs/default.toml`
- `./configs/default.json`

## uapp 配置层级

uapp 将配置分为 `uapp配置` 和 `应用配置`：

1. 先加载 `uapp配置`（公共基础配置，如数据库/Redis连接）
2. 再加载 `应用配置`（覆盖 uapp 配置中的个性化部分）

Apollo 环境变量配置：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| ApolloAddress | Apollo 地址 | 空(不使用) |
| ApolloUAppID | uapp 应用名 | uapp |
| ApolloAppId | 当前应用名 | \<app名\> |
| ApolloCluster | 集群名 | default |
