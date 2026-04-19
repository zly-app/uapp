# MsgBus — 进程内消息通知

> 导入: `github.com/zly-app/zapp/component/msgbus`

## 功能概览

- 进程内发布-订阅模式
- 支持命名主题和全局订阅
- 订阅者支持多线程并发处理
- 自动 Recover panic
- 集成 OpenTelemetry Trace 传播
- 消息无订阅者时自动丢弃，不支持历史消息

## 使用方式

### 1. 发布消息

```go
import "github.com/zly-app/zapp/component/msgbus"

// 发布到命名主题（msg 为任意类型，不需要实现 IMsgbusMessage 接口）
msgbus.Publish(ctx, "order.created", &OrderCreatedMsg{OrderID: 123})

// 消息类型定义
type OrderCreatedMsg struct {
    OrderID int64
    ctx     context.Context
}
func (m *OrderCreatedMsg) Ctx() context.Context   { return m.ctx }
func (m *OrderCreatedMsg) Topic() string          { return "order.created" }
func (m *OrderCreatedMsg) Msg() interface{}        { return m }
```

### 2. 订阅主题

```go
// 订阅命名主题
msgbus.Subscribe("order.created", 1, func(ctx context.Context, msg msgbus.IMsgbusMessage) {
    m := msg.Msg().(*OrderCreatedMsg)
    // 处理消息
})
// 参数: 主题名, 并发处理线程数, 处理函数（无返回值）
```

### 3. 全局订阅

```go
// 订阅所有主题的消息
msgbus.SubscribeGlobal(1, func(ctx context.Context, msg msgbus.IMsgbusMessage) {
    // 收到所有主题的消息
})
```

### 4. 取消订阅

```go
subID := msgbus.Subscribe("order.created", 1, handler)
msgbus.Unsubscribe("order.created", subID)

// 取消全局订阅
globalSubID := msgbus.SubscribeGlobal(1, globalHandler)
msgbus.UnsubscribeGlobal(globalSubID)
```

### 5. 关闭主题

```go
msgbus.CloseTopic("order.created")
```

## IMsgbusMessage 接口

```go
type IMsgbusMessage interface {
    Ctx() context.Context    // 消息上下文
    Topic() string           // 消息所属主题
    Msg() interface{}        // 消息内容
}
```

## 自定义消息类型

发布消息时，`msg` 参数为 `interface{}` 类型，框架会自动将其包装为 `IMsgbusMessage`。订阅者通过 `msg.Msg()` 获取原始消息：

```go
import "github.com/zly-app/zapp/component/msgbus"

// 发布任意结构体（无需实现 IMsgbusMessage 接口）
msgbus.Publish(ctx, "user.updated", &UserUpdatedMsg{UserID: 123})

// 订阅者获取原始消息
msgbus.Subscribe("user.updated", 1, func(ctx context.Context, msg msgbus.IMsgbusMessage) {
    m := msg.Msg().(*UserUpdatedMsg)
    // 处理消息
})
```

> **注意**: 虽然 `Publish` 的 `msg` 参数是 `interface{}`，但订阅者的 handler 会收到 `IMsgbusMessage` 接口，通过 `.Msg()` 获取原始消息、`.Ctx()` 获取上下文、`.Topic()` 获取主题名。

## 典型使用场景

### 场景: 事件驱动解耦

```go
// 定义事件类型（普通结构体，无需实现 IMsgbusMessage 接口）
type UserCreatedMsg struct {
    UserID int64
}

// logic 层发布事件
func (s *Service) CreateUser(ctx context.Context, user *User) error {
    err := s.dao.Create(ctx, user)
    if err != nil { return err }

    // 发布事件
    msgbus.Publish(ctx, "user.created", &UserCreatedMsg{UserID: user.ID})
    return nil
}

// handler 层订阅事件
func init() {
    msgbus.Subscribe("user.created", 2, func(ctx context.Context, msg msgbus.IMsgbusMessage) {
        m := msg.Msg().(*UserCreatedMsg)
        // 发送欢迎邮件、初始化用户数据等
    })
}
```

## 注意事项

- 消息发布是异步的，发布后立即返回
- 如果没有订阅者，消息会被丢弃
- 新订阅者不会收到历史消息
- 订阅者的 handler 中如果 panic，会被自动 Recover 并记录日志
