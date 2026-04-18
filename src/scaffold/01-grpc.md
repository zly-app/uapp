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

## 使用方式

### 1. 启用 gRPC 服务 + 网关

```go
app := uapp.NewApp("my-service",
    grpc.WithService(),        // 启用 gRPC 服务
    grpc.WithGatewayService(), // 启用 HTTP 网关
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

// 带哈希键调用(一致性哈希)
conn := grpc.GetClientConn("my-service", grpc.WithHashKey("user-123"))

// 指定目标调用
conn := grpc.GetClientConn("my-service", grpc.WithTarget("10.0.0.1:3300"))
```

### 6. 客户端 Hook

```go
// 注册全局客户端 Hook (在 init 中)
grpc.RegistryClientHook("my-hook", func(ctx context.Context, method string) (context.Context, error) {
    // 在每次 RPC 调用前执行
    return ctx, nil
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
        grpc.WithService(),        // 启用 gRPC 服务
        grpc.WithGatewayService(), // 启用 HTTP 网关
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
