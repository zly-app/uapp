# 辅助工具集

## zsingleflight — 单飞

> 导入: `github.com/zlyuancn/zsingleflight`

### 功能概览

防止缓存击穿/请求合并：对同一个 key 的并发请求，只有一个实际执行，其余等待结果。

### 使用场景

- 防止缓存失效时大量请求同时穿透到数据库
- 合并短时间内对同一资源的并发请求

> **注意**: `cache` 组件已内置 SingleFlight 支持（`singleFlight: "single"` 配置），通常无需单独使用 zsingleflight。仅在自定义缓存逻辑时可能需要。

### 使用方式

```go
import "github.com/zlyuancn/zsingleflight"

var sf zsingleflight.Group

// 执行单飞
val, err, shared := sf.Do("key", func() (interface{}, error) {
    // 只有一个请求会执行此函数
    return loadFromDB(ctx, key)
})
```

---

## rotate — 批次轮转工具

> 导入: `github.com/zlyuancn/rotate`

### 功能概览

- 批量数据收集，到达阈值或时间到期时自动轮转(回调)
- 适用于批量写入数据库、批量发送消息等场景

### 使用场景

- 批量日志写入（收集N条或超时后批量插入DB）
- 批量消息发送
- 批量指标上报

### 使用方式

```go
import "github.com/zlyuancn/rotate"

type LogItem struct {
    Level   string
    Message string
}

// 创建轮转器
r := rotate.NewRotate[LogItem](func(items []LogItem) error {
    // 批量写入数据库
    return batchInsertLogs(items)
}, rotate.WithBatchSize(100),         // 批次大小
   rotate.WithAutoRotateTime(3*time.Second),  // 自动轮转间隔
)

// 添加数据
r.Add(LogItem{Level: "info", Message: "hello"})

// 手动轮转
r.Rotate()

// 关闭（会自动轮转剩余数据）
r.Close()
```

### 典型使用场景: 系统日志批量写入

```go
package syslog

import (
    "context"
    "time"
    "github.com/zlyuancn/rotate"
    "github.com/zly-app/zapp"
    "github.com/zly-app/zapp/core"
    zappHandler "github.com/zly-app/zapp/handler"
    "my-project/client/db"
    "my-project/conf"
)

type SysLog struct {
    r *rotate.Rotate[*LogEntry]
}

type LogEntry struct {
    Level   string
    Message string
    Time    time.Time
}

var SysLog = &SysLog{}

func Init() {
    SysLog.r = rotate.NewRotate[*LogEntry](func(items []*LogEntry) error {
        return flushToDB(items)
    },
        rotate.WithBatchSize(conf.Conf.SysLogBatchSize),
        rotate.WithAutoRotateTime(time.Duration(conf.Conf.SysLogAutoRotateTimeSec)*time.Second),
    )

    // 注册应用退出前轮转
    zapp.AddHandler(zappHandler.AfterCloseService, func(app core.IApp, handlerType zappHandler.HandlerType) {
        SysLog.r.Rotate()
    })
}

func (sl *SysLog) Write(level, message string) {
    sl.r.Add(&LogEntry{Level: level, Message: message, Time: time.Now()})
}

func flushToDB(items []*LogEntry) error {
    if len(items) == 0 { return nil }
    ctx := context.Background()
    // 批量插入数据库
    // ...
    return nil
}
```

---

## zretry — 重试

> 导入: `github.com/zlyuancn/zretry`

### 功能概览

- 函数执行重试，支持重试次数、间隔、回调

### 使用方式

```go
import "github.com/zlyuancn/zretry"

err := zretry.DoRetry(
    3,                       // 总尝试次数
    1*time.Second,           // 重试间隔
    func() error {
        return doSomething()
    },
    func(nowAttemptCount, remainCount int, err error) {
        // 每次失败回调
        log.Warn("重试", zap.Int("attempt", nowAttemptCount), zap.Error(err))
    },
)
```

---

## zapp/pkg/utils — 内置通用工具

> 导入: `github.com/zly-app/zapp/pkg/utils`

### Go — 并发控制

```go
// 并行执行，等待全部完成
err := utils.Go.GoAndWait(
    func() error { return task1() },
    func() error { return task2() },
)

// 并行执行，返回 wait 函数
waitFn := utils.Go.GoRetWait(fn1, fn2)
// ... 做其他事 ...
err := waitFn()  // 等待结果

// 泛型并发查询
results, err := utils.Go.GoQuery[int64, *User](userIDs, func(id int64) (*User, error) {
    return loadUser(ctx, id)
}, true)  // true=忽略错误项
```

### Recover — Panic 恢复

```go
// 包装函数调用，捕获 panic 转 error
err := utils.Recover.WrapCall(func() error {
    panic("oops")
})
// err != nil, utils.Recover.IsRecoverError(err) == true
```

### Ctx — Context 工具

```go
// 克隆 context（保留 Values，去除 Deadline/Done/Err）
ctx2 := utils.Ctx.CloneContext(ctx)

// 用于 gpool.Go 等异步场景，防止 context 被取消
gpool.GetDefGPool().Go(func() error {
    ctx2 := utils.Ctx.CloneContext(ctx)
    return doSomething(ctx2)
}, nil)
```

### Ternary — 三元运算

```go
// 三元表达式
result := utils.Ternary.Ternary(condition, valueIfTrue, valueIfFalse)

// 短路取值（返回第一个非零值）
result := utils.Ternary.Or(os.Getenv("HOST"), "localhost").(string)
```

### Reflect — 反射工具

```go
// 判断值是否为零值
utils.Reflect.IsZero(myStruct)  // true 如果所有字段为零值
```

### Text — 通配符匹配

```go
// 通配符匹配 (? 匹配单字符, * 匹配任意)
utils.Text.IsMatchWildcard("hello.go", "h*.go")  // true
utils.Text.IsMatchWildcardAny("hello.go", "a*", "h*")  // true
```

### Trace — OpenTelemetry 链路追踪

```go
// 创建 span
span := utils.Trace.StartSpan(ctx, "my-operation",
    utils.Trace.AttrKey("key").String("value"),
)
defer utils.Trace.EndSpan(span)

// 简写
ctx, span := utils.Trace.CtxStart(ctx, "my-operation")
defer utils.Trace.CtxEnd(ctx)

// 标记错误
utils.Trace.MarkSpanAnError(span, err)

// 获取 TraceID
traceID, spanID := utils.Trace.GetOTELTraceID(ctx)
```

### GetInstance — 获取本机 IP

```go
ip := utils.GetInstance("127.0.0.1")  // 获取第一个非回环 IP，失败返回默认值
```

---

## zutils — 通用工具2

> 导入: `github.com/zlyuancn/zutils`

### AtomicValue — 泛型原子值

```go
import "github.com/zlyuancn/zutils"

// 创建
v := zutils.NewAtomic[int](0)

// 设置
v.Set(42)

// 获取
val := v.Get()  // 42
```

### Recover — Panic 恢复

```go
err := zutils.Recover.WrapCall(func() error {
    panic("oops")
})
```

### Time — 时间解析

```go
t, err := zutils.Time(time.Local).TextToTimeOfLayout("2024-01-01 00:00:00", zutils.T.Layout)
```
