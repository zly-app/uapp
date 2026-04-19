# Cron — 定时任务

> 导入: `github.com/zly-app/service/cron`

## 功能概览

- 基于 cron 表达式的定时任务调度
- 支持一次性定时任务（指定时间触发）
- 支持失败重试、超时控制、最大并发执行数
- 64 个任务堆分区，减少锁竞争
- 支持通过 gpool 限制并发
- 任务可动态启用/禁用

## 配置 (`services.cron`)

```yaml
services:
  cron:
    threadCount: 8               # 线程数 (默认 CPU*4)
    maxTaskQueueSize: 10000      # 最大任务队列
    tasks:                       # 任务列表(也可通过代码注册)
      - name: "cleanup-expired"
        expression: "0 0 3 * * *"  # 每天3点
        isOnceTrigger: false
        disable: false
        retryCount: 3
        retrySleepMs: 1000
        maxConcurrentExecuteCount: 1
        timeoutMs: 30000
```

## 使用方式

### 1. 启用 cron 服务

```go
import "github.com/zly-app/service/cron"

app := uapp.NewApp("my-service",
    cron.WithService(),  // 启用 cron 服务
)
```

### 2. 注册 cron 任务（代码方式，推荐）

```go
func init() {
    // 注册周期任务
    cron.RegistryHandler("cleanup-expired", "0 0 3 * * *", true, func(ctx cron.IContext) error {
        // 清理过期数据
        return cleanupExpired(ctx)
    })

    // 注册一次性任务
    cron.RegistryOnceHandler("one-time-task", time.Now().Add(10*time.Minute), true, func(ctx cron.IContext) error {
        // 一次性执行
        return nil
    })
}
```

### 3. 注册参数说明

```go
cron.RegistryHandler(name, expression, enable, handler)
// name: 任务名（唯一标识）
// expression: cron 表达式（6位: 秒 分 时 日 月 周）
// enable: 是否启用
// handler: 处理函数 func(cron.IContext) error
```

### 4. IContext 接口

```go
type IContext interface {
    context.Context
    ILogger                   // 嵌入日志接口
    Task() ITask              // 获取任务信息
    Meta() interface{}        // 获取元数据
    SetMeta(meta interface{}) // 设置元数据
}

type ITask interface {
    Name() string             // 任务名
    // ... 其他任务信息
}
```

## 完整示例

```go
package main

import (
    "github.com/zly-app/service/cron"
    "github.com/zly-app/uapp"
)

func init() {
    // 每5分钟同步数据
    cron.RegistryHandler("sync-data", "0 */5 * * * *", true, func(ctx cron.IContext) error {
        return syncDataFromRemote(ctx)
    })

    // 每天凌晨2点清理
    cron.RegistryHandler("daily-cleanup", "0 0 2 * * *", true, func(ctx cron.IContext) error {
        return dailyCleanup(ctx)
    })
}

func main() {
    app := uapp.NewApp("my-service",
        cron.WithService(),
    )
    defer app.Exit()
    app.Run()
}
```
