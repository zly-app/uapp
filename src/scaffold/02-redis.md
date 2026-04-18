# Redis + redis_tool — Redis 客户端与分布式锁

## Redis 客户端

> 导入: `github.com/zly-app/component/redis`

### 功能概览

- 统一客户端接口 `UniversalClient` (支持单机/集群/哨兵)
- 自动集成 OpenTelemetry Tracing 和 Metrics
- 通过 zapp 配置自动创建和生命周期管理
- 支持多实例（按名获取）

### 配置 (`components.redis.{name}`)

```yaml
components:
  redis:
    default:                          # Redis 组件名
      address: "localhost:6379"       # 地址，集群用逗号分隔
      userName: ""
      password: ""
      db: 0                           # DB 编号，仅非集群有效
      minIdle: 2                      # 最小闲置连接
      maxIdle: 4                      # 最大闲置连接
      poolSize: 10                    # 连接池大小
      idleTimeout: 3600               # 空闲超时(秒)
      waitTimeout: 5                  # 等待获取连接超时(秒)
      connectTimeout: 5               # 连接超时(秒)
      maxConnLifetime: 3600           # 连接最大存活(秒)
      maxRetries: 0                   # 操作重试次数 (<1禁用)
      readTimeoutSec: 5               # 读超时(秒)
      writeTimeoutSec: 5              # 写超时(秒)
      tlsCAFile: ""                   # TLS 根证书
      insecureSkipVerify: false       # 跳过 TLS 验证
```

### 使用方式

```go
import "github.com/zly-app/component/redis"

// 获取默认客户端
client, err := redis.GetDefClient()

// 获取命名客户端
client, err := redis.GetClient("my-redis")

// 使用 go-redis 标准 API
val, err := client.Get(ctx, "key").Result()
err = client.Set(ctx, "key", "value", time.Minute).Err()
```

### 手动初始化模式

当需要在配置加载完成后才能确定 Redis 组件名时，使用手动初始化：

```go
import (
    redis_comp "github.com/zly-app/component/redis"
    "github.com/zlyuancn/redis_tool"
)

func main() {
    // 在 NewApp 之前设置手动初始化
    redis_tool.SetManualInit()

    app := uapp.NewApp("my-service", ...)

    // 配置加载完成后手动初始化
    redis_tool.RedisClientName = conf.Conf.RedisName
    redis_tool.ManualInit()

    // 之后正常使用
    app.Run()
}
```

---

## redis_tool — Redis 分布式锁

> 导入: `github.com/zlyuancn/redis_tool`

### 功能概览

- **AutoLock**: 自动锁，返回解锁和续期函数
- **Lock/UnLock**: 手动锁，需自行管理解锁
- **RenewLock**: 锁续期
- **CheckLockCheckCode**: 授权码验证

### 初始化

redis_tool 依赖 `component/redis`，必须先初始化 Redis 组件。支持两种方式：

1. **自动初始化**: 随 zapp 框架自动初始化（默认）
2. **手动初始化**: 见上方 "手动初始化模式"

### API

```go
// 自动锁（推荐）
unlock, renew, err := redis_tool.AutoLock(ctx, lockKey, ttl)
if err == redis_tool.LockIsUsedByAnother {
    // 锁被别人持有
    return
}
if err != nil { return }
defer unlock()

// 续期
err = renew(ctx, newTtl)
```

```go
// 手动锁
checkCode, err := redis_tool.Lock(ctx, lockKey, ttl)
if err == redis_tool.LockIsUsedByAnother { return }
defer redis_tool.UnLock(ctx, lockKey, checkCode)

// 手动续期
redis_tool.RenewLock(ctx, lockKey, checkCode, newTtl)

// 验证授权码
redis_tool.CheckLockCheckCode(ctx, key, authCode)
```

### 类型

```go
type KeyUnlock = func() error                              // 解锁函数
type KeyTtlRenew = func(ctx context.Context, ttl time.Duration) error  // 续期函数
```

### 典型使用场景

```go
// 场景1: 管理操作互斥 (如防止并发修改)
func (s *Service) UpdateData(ctx context.Context, id string) error {
    unlock, _, err := redis_tool.AutoLock(ctx, fmt.Sprintf("lock:data:%s", id), 30*time.Second)
    if err == redis_tool.LockIsUsedByAnother {
        return status.Errorf(codes.Aborted, "操作冲突，请稍后重试")
    }
    if err != nil { return err }
    defer unlock()

    // 执行更新逻辑...
    return nil
}

// 场景2: 长时间运行任务 + 续期
func (p *Processor) Run(ctx context.Context) error {
    checkCode, err := redis_tool.Lock(ctx, p.lockKey, 30*time.Second)
    if err != nil { return err }

    // 启动续期协程
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        for range ticker.C {
            redis_tool.RenewLock(ctx, p.lockKey, checkCode, 30*time.Second)
        }
    }()

    // 执行长时间任务...
    defer redis_tool.UnLock(ctx, p.lockKey, checkCode)
    return nil
}
```
