# Metrics — 指标收集

> 导入: `github.com/zly-app/zapp/component/metrics`

## 功能概览

- 提供 Counter/Gauge/Histogram/Summary 四种指标类型
- 默认 Noop 实现，需配合 Prometheus 等插件注入真实实现
- 框架自动收集 gRPC/HTTP 调用的指标（通过 base.metrics filter）
- 支持自定义业务指标

## 使用方式

### 1. 注册并使用指标

```go
import "github.com/zly-app/zapp/component/metrics"

// 注册 Counter: RegistryCounter(name, help, constLabels, labels...)
metrics.RegistryCounter("orders_total", "订单总数", nil, "status")
// constLabels 为固定标签(nil 表示无)，labels 为可变标签名

// 使用 Counter — 标签直接作为方法参数传入（类型为 map[string]string）
metrics.Counter("orders_total").Inc(metrics.Labels{"status": "success"}, nil)        // +1
metrics.Counter("orders_total").Add(5, metrics.Labels{"status": "failed"}, nil)      // +5
// Inc(labels Labels, exemplar Labels)  — exemplar 通常传 nil
// Add(v float64, labels Labels, exemplar Labels)

// 注册 Gauge: RegistryGauge(name, help, constLabels, labels...)
metrics.RegistryGauge("active_connections", "活跃连接数", nil, "type")

// 使用 Gauge
metrics.Gauge("active_connections").Set(100, metrics.Labels{"type": "grpc"})
metrics.Gauge("active_connections").Inc(metrics.Labels{"type": "grpc"})
metrics.Gauge("active_connections").Dec(metrics.Labels{"type": "grpc"})
metrics.Gauge("active_connections").Add(10, metrics.Labels{"type": "grpc"})
metrics.Gauge("active_connections").Sub(5, metrics.Labels{"type": "grpc"})

// 注册 Histogram: RegistryHistogram(name, help, buckets, constLabels, labels...)
metrics.RegistryHistogram("request_duration_ms", "请求耗时",
    []float64{10, 50, 100, 500, 1000},  // buckets
    nil,                                  // constLabels
    "method",                             // labels...
)

// 使用 Histogram
metrics.Histogram("request_duration_ms").Observe(42.5, metrics.Labels{"method": "GetUser"}, nil)
// Observe(v float64, labels Labels, exemplar Labels)

// 注册 Summary: RegistrySummary(name, help, constLabels, labels...)
metrics.RegistrySummary("response_size_bytes", "响应大小", nil, "method")

// 使用 Summary
metrics.Summary("response_size_bytes").Observe(1024, metrics.Labels{"method": "GetUser"}, nil)
```

> **重要**: 指标接口**没有** `.WithLabels()` 方法。标签（`Labels = map[string]string`）直接作为每次方法调用的参数传入。`exemplar` 参数用于 Prometheus Exemplar，通常传 `nil`。

### 2. 配合 Prometheus 插件

```go
import "github.com/zly-app/plugin/prometheus"

app := uapp.NewApp("my-service",
    prometheus.WithPlugin(),  // 启用 Prometheus，自动注入 metrics.Client
)
```

### 3. 配置 Prometheus 插件

```yaml
plugins:
  metrics:
    pushGatewayAddress: ""         # PushGateway 地址
    pushInterval: 10000            # 推送间隔(毫秒)
    pushRetry: 3                   # 推送重试次数
    pushRetryInterval: 1000        # 推送重试间隔(毫秒)
    remoteWriteUrl: ""             # Prometheus Remote Write 地址
```

## 框架自动收集的指标

当 `base.metrics` filter 启用时，框架会自动收集：

- `grpc.server.started` / `grpc.server.handled` — gRPC 服务端请求计数
- `grpc.client.started` / `grpc.client.handled` — gRPC 客户端请求计数
- `http.client.started` / `http.client.handled` — HTTP 客户端请求计数
- `*.duration` — 调用耗时直方图
- `*.panic` — panic 计数
