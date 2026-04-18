# Config — 配置文件与配置变更监听

> 导入: `github.com/zly-app/zapp/config`

## 功能概览

- 多来源配置加载: 命令行 > 自定义Viper > Apollo > 默认文件
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
    err := config.Conf.Parse(ConfigKey, &Conf, true)
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

### 2. 解析组件配置

```go
// 解析指定组件配置
var redisCfg redis.RedisConfig
err := config.Conf.ParseComponentConfig("redis", "default", &redisCfg, true)
```

### 3. 配置变更监听

```go
// 监听配置变更（需要 Apollo 插件）
config.Conf.WatchKey("my-service", "someKey",
    config.WithWatchKeyOnChange(func(key string, value interface{}) {
        // 配置变更回调
        log.Info("配置变更", zap.String("key", key), zap.Any("value", value))
    }),
)
```

### 4. 注册 Apollo 命名空间

```go
// 在 main.go 中，NewApp 之前
config.RegistryApolloNeedParseNamespace("my-custom-namespace")
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
