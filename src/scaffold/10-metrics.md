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

// 注册 Counter
counter := metrics.RegistryCounter("orders_total",
    metrics.WithCounterHelp("订单总数"),
    metrics.WithCounterLabels("status"),  // 标签名
)

// 使用 Counter
metrics.Counter("orders_total", "success").Inc()        // +1
metrics.Counter("orders_total", "failed").Add(5)        // +5

// 注册 Gauge
metrics.RegistryGauge("active_connections",
    metrics.WithGaugeHelp("活跃连接数"),
    metrics.WithGaugeLabels("type"),
)

// 使用 Gauge
metrics.Gauge("active_connections", "grpc").Set(100)
metrics.Gauge("active_connections", "grpc").Inc()
metrics.Gauge("active_connections", "grpc").Dec()

// 注册 Histogram
metrics.RegistryHistogram("request_duration_ms",
    metrics.WithHistogramHelp("请求耗时"),
    metrics.WithHistogramLabels("method"),
    metrics.WithHistogramBuckets([]float64{10, 50, 100, 500, 1000}),
)

// 使用 Histogram
metrics.Histogram("request_duration_ms", "GetUser").Observe(42.5)

// 注册 Summary
metrics.RegistrySummary("response_size_bytes",
    metrics.WithSummaryHelp("响应大小"),
    metrics.WithSummaryLabels("method"),
)

// 使用 Summary
metrics.Summary("response_size_bytes", "GetUser").Observe(1024)
```

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
