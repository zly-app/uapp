# gRPC — 对外 API 网关 + 内部 gRPC 调用

> 导入: `github.com/zly-app/grpc`

## 功能概览

- **gRPC 服务端**: 启动 gRPC 服务器，自动集成 trace/metrics/日志/数据校验
- **gRPC-Gateway**: HTTP → gRPC 代理，对外提供 RESTful API
- **gRPC 客户端**: 连接池管理，支持服务发现和负载均衡
- **服务注册/发现**: 支持 static/redis 等注册中心
- **负载均衡**: round_robin / weight_random / weight_hash / weight_consistent_hash

## 配置

### 服务端配置 (`services.grpc`)

```yaml
services:
  grpc:
    bind: ":3300"                    # 监听地址
    heartbeatTime: 20                # 心跳秒数
    reqDataValidate: true            # 请求数据校验
    sendDetailedErrorInProduction: false  # 生产环境详细错误
    registryAddress: "static"        # 注册地址
    publishName: ""                  # 公告名
    publishAddress: ""               # 公告地址
    publishWeight: 100               # 公告权重
    tlsCertFile: ""                  # TLS证书
    tlsKeyFile: ""                   # TLS密钥
```

### 网关配置 (`services.grpc-gateway`)

```yaml
services:
  grpc-gateway:
    bind: ":8080"                    # HTTP网关监听地址
```

### 网关响应结构说明

使用 gRPC-Gateway 时，proto 定义的响应结构**不会**直接返回给 HTTP 客户端，而是被 `ForwardResponseRewriter`（位于 `github.com/zly-app/grpc/gateway/response.go`）自动包装为以下结构：

```json
{
  "code": 0,
  "message": "",
  "data": { /* proto 定义的原始响应结构 */ },
  "trace_id": "xxx"
}
```

| 场景 | `code` | `message` | `data` |
|------|--------|-----------|--------|
| 正常响应 | 0 | 空 | proto 定义的完整响应消息 |
| 错误响应 | gRPC 状态码 | 错误信息 | 空 |

例如，proto 定义的响应：

```protobuf
message GetUserRsp {
  string name = 1;
  int64 age = 2;
}
```

HTTP 客户端实际收到的 JSON：

```json
{
  "code": 0,
  "message": "",
  "data": {
    "name": "Alice",
    "age": 30
  },
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

### 客户端配置 (`components.grpc.{name}`)

```yaml
components:
  grpc:
    my-service:                      # 客户端名称
      address: "localhost:3300"      # 服务地址
      balance: "weight_consistent_hash"  # 均衡器
      minIdle: 2                     # 最小闲置连接
      maxIdle: 4                     # 最大闲置连接
      maxActive: 10                  # 最大活跃连接
      waitTimeout: 3                 # 等待超时(秒)
      connectTimeout: 5              # 连接超时(秒)
      idleTimeout: 3600              # 闲置超时(秒)
```

## 编译 Proto 文件

> **注意**: 本项目使用 `protoc` + Go 插件的方式编译 proto 文件，**不要使用 buf 等其他工具**。buf 等工具的配置和导入路径与本项目不一致，可能导致生成的代码无法正常工作。

### 1. 安装 protoc 编译器

从 https://github.com/protocolbuffers/protobuf/releases 下载 protoc 编译器，解压 protoc 执行文件到 `${GOPATH}/bin/`

### 2. 安装 Go 插件

```shell
# 消息类型代码生成
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# gRPC 服务代码生成
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# HTTP 网关代码生成
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Swagger 文档生成
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# 数据校验代码生成
go install github.com/envoyproxy/protoc-gen-validate@latest
```

### 3. 获取依赖 proto 文件

本项目的 proto 文件依赖 `grpc/protos` 中的公共定义（如 `google/api/annotations.proto`、`validate/validate.proto`），需将其克隆到本地：

**Linux / macOS:**

```bash
mkdir -p ${GOPATH}/protos/zly-app && cd ${GOPATH}/protos/zly-app
git clone --depth=1 https://github.com/zly-app/grpc.git
```

**Windows CMD:**

```shell
if not exist %GOPATH%\protos\zly-app mkdir %GOPATH%\protos\zly-app
cd /d %GOPATH%\protos\zly-app
git clone --depth=1 https://github.com/zly-app/grpc.git
```

**Windows PowerShell:**

```powershell
New-Item -ItemType Directory -Force -Path "$env:GOPATH\protos\zly-app" | Out-Null
cd "$env:GOPATH\protos\zly-app"
git clone --depth=1 https://github.com/zly-app/grpc.git
```

> **GoLand 用户**: 在 `设置` → `语言和框架` → `Protocol Buffers/协议缓冲区` → `Import Paths`，取消勾选 `Configure automatically/自动配置`，将 `${GOPATH}/protos/zly-app/grpc/protos` 添加到 IDE 的 proto 导入路径。

### 4. 编译命令

#### 仅基础编译（消息 + gRPC 服务）

**Linux:**

```bash
protoc \
--go_out . --go_opt paths=source_relative \
--go-grpc_out . --go-grpc_opt paths=source_relative \
pb/hello/hello.proto
```

**Windows PowerShell:**

```powershell
protoc `
--go_out . --go_opt paths=source_relative `
--go-grpc_out . --go-grpc_opt paths=source_relative `
pb/hello/hello.proto
```

生成文件：
- `hello.pb.go` — 消息类型代码
- `hello_grpc.pb.go` — gRPC 服务代码

#### 带 HTTP 网关 + 数据校验 + Swagger

**Linux:**

```bash
protoc \
-I . \
-I ${GOPATH}/protos/zly-app/grpc/protos \
--go_out . --go_opt paths=source_relative \
--go-grpc_out . --go-grpc_opt paths=source_relative \
--grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
--validate_out "lang=go:." --validate_opt paths=source_relative \
--openapiv2_out . \
pb/hello/hello.proto
```

**Windows PowerShell:**

```powershell
protoc `
-I . `
-I $env:GOPATH/protos/zly-app/grpc/protos `
--go_out . --go_opt paths=source_relative `
--go-grpc_out . --go-grpc_opt paths=source_relative `
--grpc-gateway_out . --grpc-gateway_opt paths=source_relative `
--validate_out "lang=go:." --validate_opt paths=source_relative `
--openapiv2_out . `
pb/hello/hello.proto
```

生成文件：
- `hello.pb.go` — 消息类型代码
- `hello_grpc.pb.go` — gRPC 服务代码
- `hello.pb.gw.go` — HTTP 网关代码
- `hello.pb.validate.go` — 数据校验代码
- `hello.swagger.json` — Swagger 文档

> **提示**: 使用了 `google/api/annotations.proto` 或 `validate/validate.proto` 时，必须通过 `-I` 参数指定依赖 proto 文件的路径。

### 5. Makefile 模板

```makefile
proto:
	protoc \
    -I . \
    -I ${GOPATH}/protos/zly-app/grpc/protos \
    --go_out . --go_opt paths=source_relative \
    --go-grpc_out . --go-grpc_opt paths=source_relative \
    --grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
    --validate_out "lang=go:." --validate_opt paths=source_relative \
    --openapiv2_out . \
    ./*.proto
```

## 使用方式

### 1. 启用 gRPC 服务 + 网关

```go
app := uapp.NewApp("my-service",
    grpc.WithService(),              // 启用 gRPC 服务
    grpc.WithGatewayService(),       // 启用 HTTP 网关
)
```

### 2. 定义 Proto

```protobuf
syntax = "proto3";
package myservice;

option go_package = "my-project/pb";

import "google/api/annotations.proto";
import "validate/validate.proto";

service MyService {
  rpc GetUser(GetUserReq) returns (GetUserRsp) {
    option (google.api.http) = { post: "/MyService/GetUser", body: "*" };
  }
}

message GetUserReq {
  int64 id = 1 [(validate.rules).int64.gt = 0];
}

message GetUserRsp {
  string name = 1;
}
```

### 3. 注册 gRPC 服务端

```go
// 在 main.go 中
pb.RegisterMyServiceServer(grpc.Server("my-service"), logic.NewServer())
```

> **同名服务复用 Server**: 同一个 `serverName` 的多个 gRPC 服务会自动复用同一个 `GRpcServer` 实例（共享监听端口和配置）。
> 对同一个 `serverName` 多次调用 `grpc.Server()` 会触发 panic，正确的做法是只调用一次 `grpc.Server()` 获取注册器，然后在同一个注册器上注册多个服务。
>
> ```go
> // 正确：获取一次注册器，注册多个服务（自动复用同一个 server）
> registrar := grpc.Server("my-service")
> pb.RegisterServiceAServer(registrar, implA)
> pb.RegisterServiceBServer(registrar, implB)
>
> // 错误：对同一个 serverName 多次调用 grpc.Server() 会 panic
> pb.RegisterServiceAServer(grpc.Server("my-service"), implA)
> pb.RegisterServiceBServer(grpc.Server("my-service"), implB)  // panic!
> ```

### 4. 注册 gRPC-Gateway

```go
// 在 main.go 中
client := pb.NewMyServiceClient(grpc.GetGatewayClientConn("my-service"))
_ = pb.RegisterMyServiceHandlerClient(context.Background(), grpc.GetGatewayMux(), client)
```

### 5. gRPC 客户端调用

```go
// 获取客户端连接
conn := grpc.GetClientConn("my-service")

// 普通调用
client := pb.NewMyServiceClient(conn)
rsp, err := client.GetUser(ctx, &pb.GetUserReq{Id: 1})

// 带哈希键调用(一致性哈希) — WithHashKey/WithTarget 是 grpc.CallOption，在 RPC 调用时传入
rsp, err := client.GetUser(ctx, &pb.GetUserReq{Id: 1}, grpc.WithHashKey("user-123"))

// 指定目标调用
rsp, err := client.GetUser(ctx, &pb.GetUserReq{Id: 1}, grpc.WithTarget("10.0.0.1:3300"))
```

### 6. 客户端 Hook

```go
// 注册指定服务的客户端 Hook (在 init 中)
// Hook 类型为 grpc.UnaryClientInterceptor
grpc.RegistryClientHook("my-service", func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (context.Context, error) {
    // 在每次 RPC 调用前/后执行
    return invoker(ctx, method, req, reply, cc, opts...)
})

// 给所有 client 添加 hook
grpc.RegistryAllClientHook(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (context.Context, error) {
    return invoker(ctx, method, req, reply, cc, opts...)
})
```

## 完整 main.go 示例

```go
package main

import (
    "context"

    "github.com/zly-app/grpc"
    "github.com/zly-app/uapp"
    pb "my-project/pb"
    "my-project/logic"
)

func main() {
    app := uapp.NewApp("my-service",
        grpc.WithService(),              // 启用 gRPC 服务
        grpc.WithGatewayService(),       // 启用 HTTP 网关
    )
    defer app.Exit()

    // 注册 gRPC 服务实现
    pb.RegisterMyServiceServer(grpc.Server("my-service"), logic.NewServer())

    // 注册 gRPC-Gateway 路由
    client := pb.NewMyServiceClient(grpc.GetGatewayClientConn("my-service"))
    _ = pb.RegisterMyServiceHandlerClient(context.Background(), grpc.GetGatewayMux(), client)

    app.Run()
}
```
