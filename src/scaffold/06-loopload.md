# LoopLoad — 周期加载器

> 导入: `github.com/zly-app/utils/loopload`

## 功能概览

- 泛型设计 `LoopLoad[T]`，定时从数据源加载/刷新数据到内存
- 原子值存储，并发安全读取
- 自动注册 zapp 生命周期：启动时加载，退出时关闭
- 支持手动触发刷新
- 加载操作经过 Filter 链（trace/metrics/日志）

## 使用方式

### 1. 创建 LoopLoad

```go
import "github.com/zly-app/utils/loopload"

var userLoader *loopload.LoopLoad[map[int64]*UserInfo]

func init() {
    userLoader = loopload.New[map[int64]*UserInfo]("user-loader",
        func(ctx context.Context) (map[int64]*UserInfo, error) {
            // 从数据库/配置中心加载
            users, err := loadAllUsers(ctx)
            return users, err
        },
        loopload.WithReloadTime(5*time.Minute),  // 每5分钟重载
    )
}
```

### 2. 读取数据

```go
// 获取当前加载的数据（需要传入 ctx，并发安全）
users := userLoader.Get(ctx)
if user, ok := users[userID]; ok {
    // 使用 user
}
```

### 3. 手动刷新

```go
err := userLoader.Load(ctx)
```

### 4. 选项

```go
loopload.WithReloadTime(duration)  // 设置重载间隔，默认1分钟
```

## 典型使用场景

### 场景1: 配置热更新

```go
var appConfig *loopload.LoopLoad[*AppConfig]

func init() {
    appConfig = loopload.New[*AppConfig]("app-config",
        func(ctx context.Context) (*AppConfig, error) {
            var cfg AppConfig
            err := config.Conf.Parse("my-app", &cfg, true)
            return &cfg, err
        },
        loopload.WithReloadTime(1*time.Minute),
    )
}

// 使用
cfg := appConfig.Get(ctx)
```

### 场景2: 数据预加载

```go
var dictLoader *loopload.LoopLoad[map[string]string]

func init() {
    dictLoader = loopload.New[map[string]string]("dict",
        func(ctx context.Context) (map[string]string, error) {
            return loadDictFromDB(ctx)
        },
        loopload.WithReloadTime(10*time.Minute),
    )
}

func Translate(ctx context.Context, code string) string {
    dict := dictLoader.Get(ctx)
    return dict[code]
}
```
