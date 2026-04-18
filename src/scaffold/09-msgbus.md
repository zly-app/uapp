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

// 发布到命名主题
msgbus.Publish(ctx, "order.created", &OrderCreatedMsg{OrderID: 123})

// 消息必须实现 IMsgbusMessage 接口
type OrderCreatedMsg struct {
    OrderID int64
}
func (m *OrderCreatedMsg) MsgbusTopic() string { return "order.created" }
```

### 2. 订阅主题

```go
// 订阅命名主题
msgbus.Subscribe("order.created", 1, func(ctx context.Context, msg msgbus.IMsgbusMessage) error {
    m := msg.(*OrderCreatedMsg)
    // 处理消息
    return nil
})
// 参数: 主题名, 并发处理线程数, 处理函数
```

### 3. 全局订阅

```go
// 订阅所有主题的消息
msgbus.SubscribeGlobal(1, func(ctx context.Context, msg msgbus.IMsgbusMessage) error {
    // 收到所有主题的消息
    return nil
})
```

### 4. 取消订阅

```go
subID := msgbus.Subscribe("order.created", 1, handler)
msgbus.Unsubscribe("order.created", subID)
```

### 5. 关闭主题

```go
msgbus.CloseTopic("order.created")
```

## IMsgbusMessage 接口

```go
type IMsgbusMessage interface {
    MsgbusTopic() string  // 返回消息所属主题
}
```

## 自定义消息类型

```go
import "github.com/zly-app/zapp/component/msgbus"

// 使用 SimpleMsg 快速创建消息
msg := msgbus.NewSimpleMsg("user.updated", userData)

// 或自定义结构体
type UserUpdatedMsg struct {
    UserID int64
}
func (m *UserUpdatedMsg) MsgbusTopic() string { return "user.updated" }
```

## 典型使用场景

### 场景: 事件驱动解耦

```go
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
    msgbus.Subscribe("user.created", 2, func(ctx context.Context, msg msgbus.IMsgbusMessage) error {
        m := msg.(*UserCreatedMsg)
        // 发送欢迎邮件、初始化用户数据等
        return nil
    })
}
```

## 注意事项

- 消息发布是异步的，发布后立即返回
- 如果没有订阅者，消息会被丢弃
- 新订阅者不会收到历史消息
- 订阅者的 handler 中如果 panic，会被自动 Recover 并记录日志
