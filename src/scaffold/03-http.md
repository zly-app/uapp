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

// POST JSON（body 为 []byte，通过 Option 设置请求体和响应解析）
reqData, _ := json.Marshal(&MyRequest{Name: "test"})
rsp, err := client.Post(ctx, "/api/users", reqData,
    zhttp.WithOutJson(&result),
    zhttp.WithTimeout(10*time.Second),
)

// 使用 Request 对象 + Option（推荐用于复杂请求）
var result MyResponse
rsp, err := client.Do(ctx, &zhttp.Request{
    Method: "POST",
    Path:   "/api/data",
}, zhttp.WithInJson(&data), zhttp.WithOutJson(&result))
```

### 4. 常用选项

```go
zhttp.WithTimeout(10*time.Second)      // 超时
zhttp.WithInHeader(headers)            // 请求头 (http.Header)
zhttp.WithInParams(params)             // 查询参数 (url.Values)
zhttp.WithInJson(inPtr)                // JSON 请求体（传入结构体指针，自动序列化）
zhttp.WithInYaml(inPtr)                // YAML 请求体（传入结构体指针，自动序列化）
zhttp.WithInBodyStream(reader)         // 流式请求体 (io.Reader)
zhttp.WithOutJson(outPtr)              // JSON 响应解析（传入结果指针）
zhttp.WithOutYaml(outPtr)              // YAML 响应解析（传入结果指针）
zhttp.WithOutIsStream(true)            // 流式响应（使用 rsp.BodyStream，需手动 Close）
zhttp.WithInsecureSkipVerify()         // 跳过 TLS 验证
zhttp.WithProxy("socks5://proxy:1080") // 代理
```

> **`WithInJson` vs Request.Body**: `WithInJson(inPtr)` 通过 sonic 自动将结构体序列化为 JSON 请求体；`Request.Body` 是 `string` 类型，用于直接设置原始请求体。两者可同时使用，WithInJson 优先。

### 5. Request 结构体

```go
type Request struct {
    Method            string        // HTTP 方法
    Path              string        // 请求路径
    Body              string        // 原始请求体（string 类型）
    Timeout           time.Duration // 超时
    InsecureSkipVerify bool          // 跳过 TLS 验证
    Header            Header        // 请求头 (= http.Header)
    Params            Values        // 查询参数 (= url.Values)
    InIsStream        bool          // 标记输入是流数据（库自动设置）
    OutIsStream       bool          // 标记响应是流数据
    Proxy             string        // 代理地址
}
```

> 请求体通过 `WithInJson`/`WithInYaml`/`WithInBodyStream` Option 设置（对应私有字段），而非 Request 的公开字段。

### 6. 跳过 zapp Filter

```go
// 某些场景需要跳过 zapp 的 filter（如超时控制、协程池限制）
import "github.com/zly-app/zapp/filter"

ctx = filter.WithoutFilterName(ctx, "base.timeout", "base.gpool")
```

### 7. 流式请求示例

```go
client := zhttp.NewClient("uriFileDataSource")
rsp, err := client.Do(ctx, &zhttp.Request{
    Method: "GET",
    Path:   "http://example.com/large-file",
}, zhttp.WithOutIsStream(true))
if err != nil { return err }
defer rsp.BodyStream.Close()

// 读取流
scanner := bufio.NewScanner(rsp.BodyStream)
for scanner.Scan() {
    line := scanner.Text()
    // 处理每一行
}
```

### 8. Response 结构体

```go
type Response struct {
    Body       string        // 响应体（字符串形式）
    BodyStream io.ReadCloser // 响应体流（WithOutIsStream 时使用，需手动 Close）
}
```
