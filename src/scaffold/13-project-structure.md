# 项目结构与集成指南

## 推荐项目结构

```
my-project/
├── main.go                  # 程序入口
├── go.mod
├── go.sum
├── Dockerfile
├── configs/
│   └── default.yaml         # 默认配置文件
├── conf/
│   └── config.go            # 业务配置结构定义与初始化
├── client/
│   └── db/
│       └── db.go            # 数据库/Redis客户端获取封装
├── dao/                     # 数据访问层
│   └── {table_name}/
│       └── dao.go
├── logic/                   # 业务逻辑层 (gRPC 服务实现)
│   └── {service_name}.go
├── module/                  # 业务模块 (非 gRPC 直接暴露的核心逻辑)
│   └── {module_name}.go
├── handler/                 # 事件处理器
│   └── handler.go
├── model/                   # 数据模型
│   └── model.go
├── pb/                      # protobuf 生成代码
│   ├── {service}.proto
│   ├── {service}.pb.go
│   ├── {service}_grpc.pb.go
│   └── {service}.pb.gw.go
└── assets/                  # 静态资源
```

## 各层职责说明

| 目录 | 职责 | 示例 |
|------|------|------|
| `conf/` | 配置结构定义、默认值填充、初始化 | 解析 `config.Conf.Parse()` 结果 |
| `client/db/` | 封装组件获取，对外提供统一接口 | `db.GetRedis()`, `db.GetSqlx()` |
| `dao/` | 数据访问层，封装 SQL/Redis 操作 | CRUD、批量查询 |
| `logic/` | gRPC 服务实现，组合 dao/module | 处理 RPC 请求，编排业务流程 |
| `module/` | 独立业务模块，可被 logic 或其他 module 调用 | 处理器、查询引擎、恢复器 |
| `handler/` | 事件处理器，观察者模式 | 监听业务事件，异步处理 |
| `model/` | 数据结构定义 | 请求/响应/数据库模型 |

---

## 典型 main.go 模板

### 模板: API 网关服务 (gRPC + Gateway + Redis + MySQL)

```go
package main

import (
    "context"

    "github.com/zly-app/grpc"
    "github.com/zly-app/uapp"
    "github.com/zlyuancn/redis_tool"
    "github.com/zly-app/zapp/config"
    "go.uber.org/zap"
    pb "my-project/pb"
    "my-project/conf"
    "my-project/logic"
)

func main() {
    // 如果使用 redis_tool 手动初始化
    redis_tool.SetManualInit()

    // 注册 Apollo 命名空间（如果需要）
    config.RegistryApolloNeedParseNamespace(conf.ConfigKey)

    // 创建应用
    app := uapp.NewApp("my-service",
        grpc.WithService(),              // 启用 gRPC 服务
        grpc.WithGatewayService(),       // 启用 HTTP 网关
    )
    defer app.Exit()

    // 初始化业务配置
    if err := conf.Init(); err != nil {
        zap.L().Fatal("初始化配置失败", zap.Error(err))
    }

    // 手动初始化 redis_tool
    redis_tool.RedisClientName = conf.Conf.RedisName
    redis_tool.ManualInit()

    // 注册 gRPC 服务端
    pb.RegisterMyServiceServer(grpc.Server("my-service"), logic.NewServer())

    // 注册 gRPC-Gateway
    client := pb.NewMyServiceClient(grpc.GetGatewayClientConn("my-service"))
    _ = pb.RegisterMyServiceHandlerClient(context.Background(), grpc.GetGatewayMux(), client)

    app.Run()
}
```

### 模板: 定时任务服务 (Cron + Redis + MySQL)

```go
package main

import (
    "github.com/zly-app/service/cron"
    "github.com/zly-app/uapp"
    "github.com/zly-app/zapp/config"
    "github.com/zlyuancn/redis_tool"
    "go.uber.org/zap"
    "my-project/conf"
    "my-project/handler"
    "my-project/client/db"
)

func main() {
    redis_tool.SetManualInit()
    config.RegistryApolloNeedParseNamespace(conf.ConfigKey)

    app := uapp.NewApp("my-cron-service",
        cron.WithService(),  // 启用 cron 服务
    )
    defer app.Exit()

    if err := conf.Init(); err != nil {
        zap.L().Fatal("初始化配置失败", zap.Error(err))
    }

    redis_tool.RedisClientName = conf.Conf.RedisName
    redis_tool.ManualInit()

    // 注册定时任务 (handler 内通过 db.GetSqlx()/db.GetRedis() 访问数据)
    handler.RegisterCronTasks()

    app.Run()
}
```

### 模板: 纯后台处理服务 (gpool + LoopLoad + MsgBus)

```go
package main

import (
    "context"
    "time"

    "github.com/zly-app/uapp"
    "github.com/zly-app/utils/loopload"
    "github.com/zly-app/zapp/component/gpool"
    "github.com/zly-app/zapp/component/msgbus"
    "github.com/zly-app/zapp/config"
    "github.com/zly-app/zapp/pkg/utils"
    "github.com/zlyuancn/redis_tool"
    "go.uber.org/zap"
    "my-project/conf"
    "my-project/client/db"
)

// --- LoopLoad: 周期加载数据到内存 ---
var userLoader *loopload.LoopLoad[map[int64]*UserInfo]

type UserInfo struct {
    ID   int64
    Name string
}

func init() {
    userLoader = loopload.New[map[int64]*UserInfo]("user-loader",
        func(ctx context.Context) (map[int64]*UserInfo, error) {
            // 从数据库加载全量用户数据
            return loadAllUsers(ctx)
        },
        loopload.WithReloadTime(5*time.Minute),
    )
}

// --- MsgBus: 事件订阅 ---
type UserChangedMsg struct {
    UserID int64
    ctx    context.Context
}

func (m *UserChangedMsg) Ctx() context.Context   { return m.ctx }
func (m *UserChangedMsg) Topic() string          { return "user.changed" }
func (m *UserChangedMsg) Msg() interface{}        { return m }

func init() {
    // 订阅事件，2 个并发处理线程
    msgbus.Subscribe("user.changed", 2, func(ctx context.Context, msg msgbus.IMsgbusMessage) {
        m := msg.Msg().(*UserChangedMsg)
        // 在 gpool 中异步处理，不阻塞消息总线
        ctx2 := utils.Ctx.CloneContext(ctx)
        gpool.GetDefGPool().Go(func() error {
            return processUserChanged(ctx2, m.UserID)
        }, nil)
    })
}

func main() {
    redis_tool.SetManualInit()
    config.RegistryApolloNeedParseNamespace(conf.ConfigKey)

    app := uapp.NewApp("my-worker")
    defer app.Exit()

    if err := conf.Init(); err != nil {
        zap.L().Fatal("初始化配置失败", zap.Error(err))
    }

    redis_tool.RedisClientName = conf.Conf.RedisName
    redis_tool.ManualInit()

    // 使用 LoopLoad 读取缓存数据
    ctx := context.Background()
    users := userLoader.Get(ctx)
    _ = users

    // 使用 gpool 并发处理任务
    gpool.GetDefGPool().Go(func() error {
        return runWorker(context.Background())
    }, nil)

    // 发布事件通知
    msgbus.Publish(ctx, "user.changed", &UserChangedMsg{UserID: 1, ctx: ctx})

    app.Run()
}

func loadAllUsers(ctx context.Context) (map[int64]*UserInfo, error) {
    // 从 db 加载
    return map[int64]*UserInfo{}, nil
}

func processUserChanged(ctx context.Context, userID int64) error {
    // 处理用户变更事件
    return nil
}

func runWorker(ctx context.Context) error {
    // 后台任务逻辑
    return nil
}
```

---

## 配置文件模板 (`configs/default.yaml`)

### 完整配置模板

```yaml
# 框架配置
frame:
  debug: false
  env: "production"

# 组件配置
components:
  # Redis
  redis:
    default:
      address: "localhost:6379"
      password: ""
      db: 0
      poolSize: 10

  # MySQL
  sqlx:
    default:
      driver: "mysql"
      source: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=true"
      maxIdleConns: 3
      maxOpenConns: 10

  # gRPC 客户端
  grpc:
    my-backend-service:
      address: "localhost:3300"
      balance: "weight_consistent_hash"
      minIdle: 2
      maxIdle: 4
      maxActive: 10

  # 缓存
  cache:
    default:
      compactor: "raw"
      serializer: "sonic_std"
      singleFlight: "single"
      expireSec: 300
      ignoreCacheFault: false
      cacheDB:
        type: "bigcache"
        bigCache:
          shards: 1024
          cleanTimeSec: 60

  # 协程池
  gpool:
    default:
      jobQueueSize: 100000
      threadCount: 0

# 服务配置
services:
  # gRPC 服务端
  grpc:
    bind: ":3300"
    heartbeatTime: 20
    reqDataValidate: true
    registryAddress: "static"

  # gRPC-Gateway
  grpc-gateway:
    bind: ":8080"

  # Cron 定时任务
  cron:
    threadCount: 8
    maxTaskQueueSize: 10000

# 插件配置
plugins:
  # OTLP 追踪+指标
  otlp:
    enable: true
    endpoint: "http://localhost:4318"
    insecure: true

  # Prometheus
  metrics:
    enable: false

  # PProf
  pprof:
    enable: true
    bind: ":6060"

# 业务配置 (自定义)
my-service:
  redisName: "default"
  sqlxName: "default"
  maxRetry: 3
```

---

## 各层代码模板

### conf/config.go

```go
package conf

import (
    "github.com/zly-app/zapp/config"
)

const ConfigKey = "my-service"

type Config struct {
    // 组件名
    RedisName string
    SqlxName  string

    // 业务参数
    MaxRetry        int
    CacheTtlSec     int
    LockTtlSec      int

    // ... 其他业务配置
}

var Conf = &Config{}

func Init() error {
    err := config.Conf.Parse(ConfigKey, Conf, true)
    if err != nil {
        return err
    }
    Conf.Check()
    return nil
}

func (c *Config) Check() {
    if c.RedisName == "" { c.RedisName = "default" }
    if c.SqlxName == "" { c.SqlxName = "default" }
    if c.MaxRetry == 0 { c.MaxRetry = 3 }
    if c.CacheTtlSec == 0 { c.CacheTtlSec = 300 }
    if c.LockTtlSec == 0 { c.LockTtlSec = 30 }
}
```

### client/db/db.go

```go
package db

import (
    "github.com/zly-app/component/redis"
    "github.com/zly-app/component/sqlx"
    "my-project/conf"
)

func GetRedis() (redis.UniversalClient, error) {
    return redis.GetClient(conf.Conf.RedisName)
}

func GetSqlx() sqlx.Client {
    return sqlx.GetCreator().GetClient(conf.Conf.SqlxName)
}
```

### dao/example/dao.go

> **注意**: `gendry` (`github.com/didi/gendry`) 是可选的动态 SQL 构建工具，需单独添加依赖。如不需要可替换为手写 SQL。

```go
package user

import (
    "context"
    "github.com/didi/gendry/builder"
    "my-project/client/db"
    "my-project/model"
)

func FindOne(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := db.GetSqlx().FindOne(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
    return &user, err
}

func FindList(ctx context.Context, ids []int64) ([]*model.User, error) {
    var users []*model.User
    // 使用 gendry 构建 IN 查询
    where := map[string]interface{}{"id": ids}
    cond, vals, _ := builder.BuildSelect("users", where, []string{"*"})
    err := db.GetSqlx().Find(ctx, &users, cond, vals...)
    return users, err
}

func Create(ctx context.Context, user *model.User) error {
    _, err := db.GetSqlx().Exec(ctx,
        "INSERT INTO users (name, age) VALUES (?, ?)",
        user.Name, user.Age,
    )
    return err
}
```

### logic/server.go (gRPC 服务实现)

```go
package logic

import (
    "context"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "github.com/zly-app/zapp/component/gpool"
    "github.com/zly-app/zapp/pkg/utils"
    pb "my-project/pb"
    "my-project/dao/user"
    "my-project/model"
    "my-project/client/db"
)

type Server struct {
    pb.UnimplementedMyServiceServer
}

func NewServer() pb.MyServiceServer {
    return &Server{}
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.GetUserRsp, error) {
    if req.Id <= 0 {
        return nil, status.Errorf(codes.InvalidArgument, "id must be positive")
    }

    result, err := user.FindOne(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
    }

    return &pb.GetUserRsp{Name: result.Name}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRsp, error) {
    // 主逻辑同步执行
    err := user.Create(ctx, &model.User{Name: req.Name, Age: req.Age})
    if err != nil {
        return nil, status.Errorf(codes.Internal, "创建失败: %v", err)
    }

    // 非关键操作异步执行
    ctx2 := utils.Ctx.CloneContext(ctx)
    gpool.GetDefGPool().Go(func() error {
        // 异步清缓存、发通知等
        rdb, _ := db.GetRedis()
        if rdb != nil {
            rdb.Del(ctx2, "user:list")
        }
        return nil
    }, nil)

    return &pb.CreateUserRsp{}, nil
}
```

### handler/handler.go (事件处理器 — 观察者模式)

```go
package handler

import (
    "context"
    "github.com/zly-app/zapp/component/gpool"
    "github.com/zly-app/zapp/component/msgbus"
    "github.com/zly-app/zapp/pkg/utils"
)

// 事件类型
type HandlerType int

const (
    AfterCreateUser HandlerType = iota
    AfterUpdateUser
    AfterDeleteUser
)

// 事件信息
type Info struct {
    HandlerType HandlerType
    UserID      int64
    Data        interface{}
    ctx         context.Context
}

// 注意: msgbus.Publish 的第三个参数为 interface{}，无需实现 IMsgbusMessage 接口
// 这里实现 IMsgbusMessage 是为了订阅者能通过 msg.Msg() 获取类型断言后的原始消息
func (i *Info) Ctx() context.Context   { return i.ctx }
func (i *Info) Topic() string          { return "handler.event" }
func (i *Info) Msg() interface{}        { return i }

var handlers = map[HandlerType][]func(ctx context.Context, info *Info) error{}

// 注册处理器
func AddHandler(ht HandlerType, fn func(ctx context.Context, info *Info) error) {
    handlers[ht] = append(handlers[ht], fn)
}

// 触发事件（异步执行所有注册的 handler）
func Trigger(ctx context.Context, ht HandlerType, info *Info) {
    info.HandlerType = ht
    info.ctx = ctx
    fns, ok := handlers[ht]
    if !ok { return }

    for _, fn := range fns {
        fn := fn
        ctx2 := utils.Ctx.CloneContext(ctx)
        gpool.GetDefGPool().Go(func() error {
            return fn(ctx2, info)
        }, nil)
    }
}
```

### handler 使用示例

```go
// 在 init 或业务初始化时注册 handler
func init() {
    handler.AddHandler(handler.AfterCreateUser, func(ctx context.Context, info *handler.Info) error {
        // 创建用户后: 发送欢迎邮件、初始化数据等
        return sendWelcomeEmail(ctx, info.UserID)
    })
}

// 在 logic 层触发事件
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRsp, error) {
    // ... 创建用户 ...

    // 触发事件
    handler.Trigger(ctx, handler.AfterCreateUser, &handler.Info{UserID: userID})

    return &pb.CreateUserRsp{}, nil
}
```

---

## Protobuf 生成命令

### 安装工具

```bash
# 安装 protoc 编译器
# 参考: https://grpc.io/docs/protoc-installation/

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
```

### 生成命令

```bash
protoc \
  -I . \
  -I third_party/googleapis \
  -I third_party/protoc-gen-validate \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
  --validate_out=. --validate_opt=paths=source_relative \
  pb/{service}.proto
```

---

## Dockerfile 模板

```dockerfile
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server .

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/server .
COPY configs/ ./configs/

EXPOSE 3300 8080

ENTRYPOINT ["./server"]
```

---

## 单独集成某个工具的指南

### 只加 Redis

1. 在 `configs/default.yaml` 添加 `components.redis.default` 配置
2. 在 `client/db/db.go` 添加 `GetRedis()` 方法
3. 代码中使用 `redis.GetClient("default")` 获取客户端

### 只加 MySQL

1. 在 `configs/default.yaml` 添加 `components.sqlx.default` 配置
2. 在 `client/db/db.go` 添加 `GetSqlx()` 方法
3. 代码中使用 `sqlx.GetCreator().GetClient("default")` 获取客户端

### 只加 Cron 定时任务

1. 在 `main.go` 的 `uapp.NewApp()` 中添加 `cron.WithService()`
2. 在 `configs/default.yaml` 添加 `services.cron` 配置
3. 用 `cron.RegistryHandler()` 注册任务

### 只加 Cache 缓存

1. 在 `configs/default.yaml` 添加 `components.cache.default` 配置
2. 代码中使用 `cache.GetCacheCreator().GetDefCache().Get/Set/Del`

### 只加 gRPC 服务

1. 定义 proto 文件并生成代码
2. 在 `main.go` 添加 `grpc.WithService()` 和 `grpc.WithGatewayService()`
3. 在 `configs/default.yaml` 添加 `services.grpc` 和 `services.grpc-gateway` 配置
4. 注册 gRPC 服务端和 Gateway

### 只加 LoopLoad

1. 在代码中创建 `loopload.New[T]()` 实例
2. 自动随 zapp 生命周期启停，无需额外配置

### 只加 gpool

- gpool 是 zapp 内置组件，默认自动创建
- 直接使用 `gpool.GetDefGPool().Go()` 即可
- 如需自定义配置，在 `configs/default.yaml` 添加 `components.gpool.default`

### 只加 MsgBus

- MsgBus 是 zapp 内置组件，默认自动创建
- 直接使用 `msgbus.Publish()` / `msgbus.Subscribe()` 即可

### 只加 Metrics

1. 默认为 Noop 实现
2. 添加 `prometheus.WithPlugin()` 启用 Prometheus 指标收集
3. 在 `configs/default.yaml` 添加 `plugins.metrics` 配置
4. 代码中使用 `metrics.Counter/Gauge/Histogram/Summary`

### 只加 redis_tool 分布式锁与原子操作

1. 确保已配置 Redis 组件
2. 在 `main.go` 的 `uapp.NewApp()` **之前**调用 `redis_tool.SetManualInit()`
3. 在配置加载完成后调用 `redis_tool.RedisClientName = "default"; redis_tool.ManualInit()`
4. 代码中使用 `redis_tool.AutoLock()`/`redis_tool.Lock()` 或 `redis_tool.CompareAndSwap()`/`redis_tool.CompareAndDel()`/`redis_tool.CompareAndExpire()`

---

## 常见组合方案

### 组合1: API 服务 (gRPC + Redis + MySQL + Cache)

```go
app := uapp.NewApp("my-api",
    grpc.WithService(),
    grpc.WithGatewayService(),
)
```

配置需要: `components.redis`, `components.sqlx`, `components.cache`, `services.grpc`, `services.grpc-gateway`

### 组合2: 数据处理服务 (gpool + Redis + LoopLoad + Cron)

```go
import (
    "time"
    "github.com/zly-app/service/cron"
    "github.com/zly-app/utils/loopload"
)

// LoopLoad 在代码中创建，自动随 zapp 生命周期启停
var dataLoader = loopload.New[map[string]string]("data-loader",
    func(ctx context.Context) (map[string]string, error) {
        return loadDataFromDB(ctx)
    },
    loopload.WithReloadTime(5*time.Minute),
)

app := uapp.NewApp("my-processor",
    cron.WithService(),
)
```

配置需要: `components.redis`, `components.sqlx`, `components.gpool`, `services.cron`

### 组合3: 事件驱动服务 (MsgBus + gpool + Cache)

```go
import (
    "github.com/zly-app/zapp/component/msgbus"
    "github.com/zly-app/zapp/component/gpool"
)

// MsgBus 和 gpool 均为内置组件，无需 WithService 选项
// 订阅事件 (在 init 或业务初始化时)
msgbus.Subscribe("order.created", 2, func(ctx context.Context, msg msgbus.IMsgbusMessage) {
    // 处理事件
})

app := uapp.NewApp("my-event-service")
```

配置需要: `components.cache`, `components.gpool`

### 组合4: 完整可观测性服务 (gRPC + Metrics + Trace)

```go
import (
    "github.com/zly-app/grpc"
    "github.com/zly-app/plugin/prometheus"
)

app := uapp.NewApp("my-service",
    grpc.WithService(),
    grpc.WithGatewayService(),
    prometheus.WithPlugin(),
)
```

配置需要: `plugins.metrics`, `plugins.otlp`(或jaeger/zipkin)
