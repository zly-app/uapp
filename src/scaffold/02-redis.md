# Redis + redis_tool — Redis 客户端与分布式锁与原子操作

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

## redis_tool — Redis 分布式锁与原子操作

> 导入: `github.com/zlyuancn/redis_tool`

### 功能概览

- **AutoLock**: 自动锁，返回解锁和续期函数
- **Lock/UnLock**: 手动锁，需自行管理解锁
- **RenewLock**: 锁续期
- **CheckLockCheckCode**: 授权码验证
- **CompareAndSwap**: CAS 原子交换
- **CompareAndDel**: 条件删除（值匹配才删除）
- **CompareAndExpire**: 条件设置过期（值匹配才设置 TTL）
- **GetRedis**: 获取 Redis 客户端

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

```go
// CAS 原子交换: key 的值从 v1 变为 v2
ok, err := redis_tool.CompareAndSwap(ctx, key, oldValue, newValue)

// 条件删除: key 的值等于 value 时删除
ok, err := redis_tool.CompareAndDel(ctx, key, value)

// 条件设置过期: key 的值等于 value 时设置 TTL
ok, err := redis_tool.CompareAndExpire(ctx, key, value, 10*time.Minute)
```

```go
// 获取 Redis 客户端（与 component/redis 共享同一实例）
client, err := redis_tool.GetRedis()
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

// 场景3: CAS 原子交换 (如状态流转)
func (s *Service) TransitionState(ctx context.Context, id string) error {
    // 仅当状态为 "pending" 时才切换为 "processing"
    ok, err := redis_tool.CompareAndSwap(ctx, fmt.Sprintf("state:%s", id), "pending", "processing")
    if err != nil { return err }
    if !ok {
        return fmt.Errorf("状态不是 pending，无法切换")
    }

    // 执行处理逻辑...
    return nil
}

// 场景4: 条件删除 (如清理特定值的缓存标记)
func (s *Service) ClearFlag(ctx context.Context, id string) error {
    ok, err := redis_tool.CompareAndDel(ctx, fmt.Sprintf("flag:%s", id), "active")
    if err != nil { return err }
    if !ok {
        // 值不匹配或 key 不存在，无需删除
        return nil
    }
    return nil
}
```
