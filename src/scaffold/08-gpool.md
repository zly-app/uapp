# GPool — 协程池

> 导入: `github.com/zly-app/zapp/component/gpool`

## 功能概览

- Worker-Pool 模型，控制 goroutine 数量
- 支持多命名池，按需从配置创建
- 自动 Recover panic
- 集成 OpenTelemetry Trace 传播
- 支持同步/异步/非阻塞/批量并发执行

## 配置 (`components.gpool.{name}`)

```yaml
components:
  gpool:
    default:
      jobQueueSize: 100000      # 任务队列大小 (最小100000)
      threadCount: 0            # worker 数: 正数=固定, 0=CPU*2, 负数=无池模式
```

## 使用方式

### 1. 获取协程池

```go
import "github.com/zly-app/zapp/component/gpool"

// 获取默认池
pool := gpool.GetDefGPool()

// 获取命名池
pool := gpool.GetGPool("my-pool")
```

### 2. 异步执行

```go
// 提交异步任务（队列满时阻塞）
gpool.GetDefGPool().Go(func() error {
    // 异步逻辑
    return nil
}, nil)  // 第二个参数为完成回调，可为 nil
```

### 3. 异步执行 + 回调

```go
gpool.GetDefGPool().Go(func() error {
    // 异步逻辑
    return result, nil
}, func(err error) {
    // 完成回调
    if err != nil { log.Error("任务失败", zap.Error(err)) }
})
```

### 4. 非阻塞提交

```go
// 队列满时返回 false 而非阻塞
ok := gpool.GetDefGPool().TryGo(func() error {
    return nil
}, nil)
if !ok {
    // 队列已满
}
```

### 5. 同步执行

```go
err := gpool.GetDefGPool().GoSync(func() error {
    // 同步逻辑
    return nil
})
```

### 6. 批量并发执行

```go
err := gpool.GetDefGPool().GoAndWait(
    func() error { return task1() },
    func() error { return task2() },
    func() error { return task3() },
)
// 等待所有完成，返回第一个非 nil error
```

## 典型使用场景

### 场景1: gRPC 方法中异步操作

```go
func (s *Server) CreateRecord(ctx context.Context, req *pb.CreateReq) (*pb.CreateRsp, error) {
    // 主逻辑同步执行
    err := s.dao.Create(ctx, record)
    if err != nil { return nil, err }

    // 非关键操作异步执行
    ctx2 := utils.Ctx.CloneContext(ctx)  // 克隆 context，避免取消
    gpool.GetDefGPool().Go(func() error {
        // 异步写历史、清缓存等
        return s.dao.WriteHistory(ctx2, record)
    }, nil)

    return &pb.CreateRsp{}, nil
}
```

### 场景2: 批量并发查询

```go
var wg sync.WaitGroup
var mu sync.Mutex
results := make([]Item, 0, len(ids))

for _, id := range ids {
    id := id
    gpool.GetDefGPool().Go(func() error {
        item, err := loadItem(ctx, id)
        if err != nil { return err }
        mu.Lock()
        results = append(results, item)
        mu.Unlock()
        return nil
    }, nil)
}
```

## 注意事项

- 在 gpool.Go 中使用 context 时，必须用 `utils.Ctx.CloneContext(ctx)` 克隆，因为原始 context 可能随请求结束而取消
- gpool.Go 中的 panic 会被自动 Recover，不会影响其他任务
