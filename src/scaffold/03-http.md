# HTTP 客户端 — 调用外部 HTTP 服务

> 导入: `github.com/zly-app/component/http`

## 功能概览

- 声明式 HTTP 客户端，自动集成 trace/metrics/日志
- 支持 JSON/YAML/Stream 等多种请求/响应格式
- 支持代理(HTTP/HTTPS/SOCKS5)、TLS、超时控制
- 可替换标准库 `http.DefaultClient` 和 `http.DefaultTransport`
- 通过 Filter 链统一处理请求/响应

## 使用方式

### 1. 替换标准库 (uapp 默认已启用)

```go
// uapp.NewApp() 内部已自动调用，无需手动
http.ReplaceStd()
```

### 2. 创建客户端

```go
import zhttp "github.com/zly-app/component/http"

client := zhttp.NewClient("my-http-client")
```

### 3. 发起请求

```go
// GET 请求
rsp, err := client.Get(ctx, "/api/users")

// POST JSON
req := &MyRequest{Name: "test"}
rsp, err := client.Post(ctx, "/api/users", req,
    zhttp.WithInJson(),
    zhttp.WithOutJson(&result),
    zhttp.WithTimeout(10*time.Second),
)

// 使用 Request 对象
rsp, err := client.Do(&zhttp.Request{
    Method: "POST",
    Path:   "/api/data",
    Body:   zhttp.WithInJsonBody(data),
    Out:    zhttp.WithOutJsonResult(&result),
})
```

### 4. 常用选项

```go
zhttp.WithTimeout(10*time.Second)      // 超时
zhttp.WithInHeader(headers)            // 请求头
zhttp.WithInParams(params)             // 查询参数
zhttp.WithInJson()                     // JSON 请求体
zhttp.WithInYaml()                     // YAML 请求体
zhttp.WithOutJson(&result)             // JSON 响应解析
zhttp.WithOutYaml(&result)             // YAML 响应解析
zhttp.WithOutIsStream(true)            // 流式响应
zhttp.WithInsecureSkipVerify(true)     // 跳过 TLS 验证
zhttp.WithProxy("socks5://proxy:1080") // 代理
```

### 5. 跳过 zapp Filter

```go
// 某些场景需要跳过 zapp 的 filter（如超时控制、协程池限制）
import "github.com/zly-app/zapp/filter"

ctx = filter.WithoutFilterName(ctx, "base.timeout", "base.gpool")
```

### 6. 流式请求示例

```go
client := zhttp.NewClient("uriFileDataSource")
rsp, err := client.Do(ctx, &zhttp.Request{
    Method: "GET",
    Path:   "http://example.com/large-file",
}, zhttp.WithOutIsStream(true))
if err != nil { return err }
defer rsp.Body.Close()

// 读取流
scanner := bufio.NewScanner(rsp.Body)
for scanner.Scan() {
    line := scanner.Text()
    // 处理每一行
}
```
