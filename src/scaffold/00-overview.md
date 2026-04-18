# uapp 脚手架文档 — 总览

> 本文档面向 AI，用于指导基于 uapp (zapp) 框架的 Go 项目脚手架生成。

## 工具能力速查表

根据项目需求选择合适的工具：

| 需求 | 工具 | 导入路径 | 详见 |
|------|------|----------|------|
| 对外提供 API 网关 + 内部 gRPC 调用 | grpc | `github.com/zly-app/grpc` | [01-grpc.md](01-grpc.md) |
| Redis 客户端 | redis | `github.com/zly-app/component/redis` | [02-redis.md](02-redis.md) |
| Redis 分布式锁/续期/授权码验证 | redis_tool | `github.com/zlyuancn/redis_tool` | [02-redis.md](02-redis.md) |
| 调用外部 HTTP 服务 | http | `github.com/zly-app/component/http` | [03-http.md](03-http.md) |
| MySQL/PostgreSQL 等 SQL 数据库 | sqlx | `github.com/zly-app/component/sqlx` | [04-sqlx.md](04-sqlx.md) |
| 本地缓存/多级缓存(进程内+Redis) | cache | `github.com/zly-app/cache/v2` | [05-cache.md](05-cache.md) |
| 周期加载/定时从数据源刷新数据 | loopload | `github.com/zly-app/utils/loopload` | [06-loopload.md](06-loopload.md) |
| 定时任务(Cron) | cron | `github.com/zly-app/service/cron` | [07-cron.md](07-cron.md) |
| 协程池 | gpool | `github.com/zly-app/zapp/component/gpool` | [08-gpool.md](08-gpool.md) |
| 进程内消息通知(发布-订阅) | msgbus | `github.com/zly-app/zapp/component/msgbus` | [09-msgbus.md](09-msgbus.md) |
| 指标收集(Counter/Gauge/Histogram) | metrics | `github.com/zly-app/zapp/component/metrics` | [10-metrics.md](10-metrics.md) |
| 配置文件/配置变更监听 | config | `github.com/zly-app/zapp/config` | [11-config.md](11-config.md) |
| 单飞(防缓存击穿) | zsingleflight | `github.com/zlyuancn/zsingleflight` | [12-utils.md](12-utils.md) |
| 批次轮转(批量写入/轮转刷新) | rotate | `github.com/zlyuancn/rotate` | [12-utils.md](12-utils.md) |
| 重试 | zretry | `github.com/zlyuancn/zretry` | [12-utils.md](12-utils.md) |
| 通用工具(并发/Recover/三元/Trace等) | zapp/pkg/utils | `github.com/zly-app/zapp/pkg/utils` | [12-utils.md](12-utils.md) |
| 通用工具2(原子值/时间解析等) | zutils | `github.com/zlyuancn/zutils` | [12-utils.md](12-utils.md) |

## 功能关键词索引

| 关键词 | 推荐工具 |
|--------|----------|
| HTTP API / RESTful / 网关 | grpc (grpc-gateway) |
| gRPC / protobuf / RPC | grpc |
| Redis / 缓存 / KV | redis + cache |
| 分布式锁 / 互斥 / 续期 | redis_tool |
| MySQL / PostgreSQL / SQL / 数据库 | sqlx |
| HTTP 客户端 / 调用外部服务 / 下载 | http |
| 本地缓存 / 进程内缓存 / 多级缓存 | cache |
| 定时加载 / 配置刷新 / 数据预热 | loopload |
| 定时任务 / Cron / 周期执行 | cron |
| 协程 / 并发 / 异步 / goroutine | gpool |
| 消息 / 事件 / 发布订阅 / 解耦 | msgbus |
| 指标 / 监控 / Counter / Histogram | metrics + prometheus |
| 配置 / 热更新 / Apollo | config + apollo_provider |
| 单飞 / 防击穿 / 去重 | zsingleflight (或 cache 内置) |
| 批量 / 轮转 / 批写入 | rotate |
| 重试 / Retry | zretry |
| 并发等待 / GoAndWait | zapp/pkg/utils.Go |
| Recover / Panic 捕获 | zapp/pkg/utils.Recover |
| Context 克隆 | zapp/pkg/utils.Ctx |
| 链路追踪 / Trace / Span | zapp/pkg/utils.Trace + otlp 插件 |
| 原子值 / AtomicValue | zutils |
| 三元 / 默认值 | zapp/pkg/utils.Ternary |
| 通配符匹配 | zapp/pkg/utils.Text |

## 按场景选择工具

### 场景: API 服务 (对外 HTTP + 内部 gRPC)
- **grpc** — 同时提供 gRPC 服务和 HTTP 网关(gRPC-Gateway)
- **config** — 解析业务配置
- **gpool** — 异步任务
- 可选: **redis**/**sqlx**/**cache** 作为数据层

### 场景: 数据处理/后台任务
- **gpool** — 并发处理
- **cron** — 定时调度
- **loopload** — 周期加载配置/数据
- **redis_tool** — 分布式锁防并发

### 场景: 缓存服务
- **cache** — 多级缓存 (进程内 + Redis)
- **redis** — Redis 客户端
- **loopload** — 周期预热

### 场景: 事件驱动
- **msgbus** — 进程内消息总线
- **gpool** — 异步处理事件

### 场景: 监控可观测性
- **metrics** — 指标收集
- **zapp/pkg/utils** 的 Trace — 链路追踪
- 配合 plugin: otlp/prometheus/zipkin

## 项目结构指南

详见 [13-project-structure.md](13-project-structure.md)

## 插件系统

uapp 默认启用的插件：
- **pprof** — Go 性能分析 HTTP 端点
- **otlp** — OpenTelemetry 追踪+指标上报

其他可用插件（需手动启用）：

| 插件 | 启用方式 | 功能 |
|------|----------|------|
| otlp | uapp 默认启用 | OTLP 协议 Trace + Metrics 上报 |
| pprof | uapp 默认启用 | Go pprof 性能分析端点 |
| prometheus | `prometheus.WithPlugin()` | Prometheus 指标采集/Remote Write |
| zipkin | `zipkin.WithPlugin()` | Zipkin 链路追踪 |
| jaegerotel | `jaegerotel.WithPlugin()` | Jaeger OpenTelemetry 链路追踪 |
| honey | `honey.WithPlugin()` | 日志采集/旋转(支持 loki/http/std) |
| apollo_provider | uapp 自动启用 | Apollo 配置观察提供者(有 apollo 配置时自动开启) |

### 插件配置

```yaml
plugins:
  otlp:
    enable: true
    endpoint: "http://localhost:4318"
    insecure: true
  pprof:
    enable: true
    bind: ":6060"
  metrics:         # prometheus
    enable: false
    pushGatewayAddress: ""
    remoteWriteUrl: ""
```

## 核心设计模式

所有 zly-app 生态组件遵循一致的设计模式：

1. **Creator + AnyConn**: 组件通过 `Creator` 接口 + `AnyConn` 连接管理器实现按名获取、单例缓存、自动关闭
2. **Config + Check()**: 每个组件都有配置结构体和 `Check()` 方法，提供合理默认值并校验
3. **Filter 链**: 所有 I/O 操作都经过 `filter.GetClientFilter` 链，统一注入 trace、metrics、日志
4. **Option 函数选项**: 使用 `WithXxx()` 函数选项模式配置行为
5. **错误占位**: 创建失败时返回 `errClient`/`errCache`/`errConn`，所有方法返回错误，避免 nil panic
6. **zapp 生命周期**: 通过 `zapp.AddHandler()` 注册初始化/关闭回调

## 应用生命周期

```
NewApp(name, opts...)
  ├── 加载配置 (config)
  ├── 初始化 Logger
  ├── 构建组件 (component: gpool + msgbus + 自定义)
  ├── 构建插件 (plugin: 按依赖拓扑排序)
  ├── 构建过滤器 (filter: 注册 → 构建链 → 初始化)
  ├── 构建服务 (service: grpc/cron 等)
  ↓
app.Run()
  ├── 启动插件 (依赖排序)
  ├── 启动服务 (WaitRun 两阶段等待)
  ├── 阻塞等待退出信号
  ↓
app.Exit() 或收到退出信号
  ├── 关闭服务
  ├── 关闭插件 (逆序)
  └── 释放组件
```

生命周期钩子（通过 `zapp.AddHandler(handlerType, fn)` 注册）：

| 钩子 | 说明 |
|------|------|
| `BeforeInitialize` | 初始化前 |
| `AfterInitialize` | 初始化后 |
| `BeforeStart` | 启动前 |
| `AfterStart` | 启动后 |
| `BeforeCloseService` | 关闭服务前 |
| `AfterCloseService` | 关闭服务后 |
| `BeforeClosePlugin` | 关闭插件前 |
| `AfterClosePlugin` | 关闭插件后 |
| `BeforeExit` | 退出前 |
| `AfterExit` | 退出后 |
