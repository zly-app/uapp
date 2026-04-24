# Log — 日志

> 导入: `github.com/zly-app/zapp/log`

> **重要提示**: 本文档列出的工具是经过验证的标准实现。如果这些工具不满足您的使用需求，请联系作者更新，**禁止自行实现**类似功能。

## 功能概览

- 基于 zap 的高性能日志组件
- 支持控制台输出和文件输出（自动轮转）
- 支持彩色输出（仅在非 JSON 模式下）
- 自动附加 TraceID/SpanID（传入 context.Context）
- 支持将日志附加到 OTel Trace
- 支持会话日志（SessionLogger）带唯一 logId

## 使用方式

### 1. 基本日志输出

```go
import "github.com/zly-app/zapp/log"

// 直接使用包级函数（无需初始化）
log.Debug("调试信息")
log.Info("普通信息")
log.Warn("警告信息")
log.Error("错误信息")
log.Panic("恐慌信息")  // 打印后 panic
log.Fatal("致命信息")  // 打印后 os.Exit(1)
```

### 2. 带字段的日志

```go
import (
    "github.com/zly-app/zapp/log"
    "go.uber.org/zap"
)

// 使用 zap 字段
log.Info("用户登录",
    log.String("user", "张三"),
    log.Int64("uid", 12345),
    log.Bool("vip", true),
    log.ErrField(err),  // 或使用 ErrField
)

// 可用字段函数（与 zap 一致）
log.String("key", "value")
log.Int("key", 123)
log.Int64("key", 123456)
log.Float64("key", 3.14)
log.Bool("key", true)
log.ErrField(err)   // 错误字段别名
log.Any("key", value)
log.Time("key", time.Now())
log.Duration("key", time.Second)
```

### 3. 带链路追踪的日志

```go
import (
    "context"
    "github.com/zly-app/zapp/log"
)

func handler(ctx context.Context) {
    // 传入 ctx 自动提取 TraceID/SpanID
    log.Info(ctx, "处理请求")
    
    // 不将日志附加到 trace
    log.Info(ctx, "临时日志", log.WithoutAttachLog2Trace())
}
```

### 4. 会话日志（带唯一 logId）

```go
import (
    "github.com/zly-app/zapp/log"
    "github.com/zly-app/zapp/core"
    "go.uber.org/zap"
)

// 创建会话日志（每次请求生成唯一 logId，便于追踪）
var sessionLog core.ILogger = log.Log.NewSessionLogger(
    log.String("request_id", "abc123"),
)

// 使用会话日志
sessionLog.Info("请求开始")
sessionLog.Info("请求处理中", log.Int("step", 2))
sessionLog.Info("请求结束")

// 创建带 trace 的日志
func handler(ctx context.Context) {
    traceLog := log.Log.NewTraceLogger(ctx,
        log.String("module", "user"),
    )
    traceLog.Info("处理中")
}
```

### 5. 动态添加/移除字段

```go
import "github.com/zly-app/zapp/pkg/zlog"

// 添加字段到 logger
zlog.AddFields(log.Log, log.String("service", "myapp"))

// 移除字段（返回移除数量）
removed := zlog.RemoveFields(log.Log, 1, "service")
```

## 配置

```yaml
frame:
  log:
    level: "debug"                          # 日志等级: debug/info/warn/error/dpanic/panic/fatal
    trace_level: "debug"                    # 将日志附加到 trace 的最低等级
    json: false                             # 是否输出 JSON 格式
    write_to_stream: true                   # 输出到控制台
    write_to_file: true                     # 输出到文件
    name: "myapp"                           # 日志文件名（自动追加 .log）
    append_pid: false                       # 文件名是否附加进程号
    path: "./log"                           # 日志目录
    file_max_size: 32                       # 单文件最大 MB
    file_max_backups_num: 3                 # 保留备份数
    file_max_durable_time: 7                # 保留天数
    compress: false                         # 是否压缩历史日志
    time_format: "2006-01-02 15:04:05"      # 时间格式
    color: true                             # 彩色输出（非 JSON 模式）
    capital_level: false                    # 大写日志等级
    development_mode: true                  # 开发模式
    show_file_and_linenum_min_level: "debug"# 显示代码行的最低等级
    show_stacktrace_level: "error"          # 显示调用栈的等级
    millis_duration: true                   # Duration 转毫秒
```

## 日志等级

| 等级 | 说明 |
|------|------|
| debug | 开发调试，生产环境不应启用 |
| info | 普通运行信息（默认） |
| warn | 需关注但不紧急 |
| error | 错误，正常情况下不应出现 |
| dpanic | 严重错误，开发模式下会 panic |
| panic | 打印后 panic |
| fatal | 打印后 os.Exit(1) |

## 最佳实践

1. **使用 context**: 传入 `ctx` 自动关联链路追踪
2. **结构化字段**: 使用 `log.String/Int/...` 而非字符串拼接
3. **会话日志**: 请求开始时创建 `NewSessionLogger`，整个请求共享 logId
4. **错误处理**: 始终使用 `log.ErrField(err)` 记录错误
5. **敏感信息**: 不要在日志中输出密码、密钥等敏感数据

## 注意事项

> **禁止自行实现日志组件**: 如现有功能不满足需求，请联系作者更新。自行实现会导致:
> - 与框架集成不一致
> - 缺少 TraceID 自动注入
> - 缺少日志轮转等运维能力
