# Cache — 本地缓存/多级缓存

> 导入: `github.com/zly-app/cache/v2`

## 功能概览

- **多级缓存后端**: bigcache(进程内) / freecache(进程内) / redis(分布式) / no_cache(无缓存)
- **缓存穿透保护**: SingleFlight 防止缓存击穿（同一 key 只加载一次）
- **缓存故障容错**: 可配置忽略缓存故障，降级从 LoadFn 加载
- **压缩/序列化**: 支持 zstd/gzip 压缩，支持 sonic/jsoniter/msgpack/yaml 序列化
- **Get 流程**: 查缓存 → 命中返回 → 未命中 → SingleFlight 调用 LoadFn → 写入缓存

## 配置 (`components.cache.{name}`)

```yaml
components:
  cache:
    default:
      compactor: "raw"              # 压缩器: raw/zstd/gzip
      serializer: "sonic_std"       # 序列化器: sonic/sonic_std/msgpack/jsoniter/json/yaml/bytes
      singleFlight: "single"        # 单跑模块: single/no
      expireSec: 300                # 默认过期时间(秒)
      ignoreCacheFault: false       # 忽略缓存故障
      cacheDB:
        type: "bigcache"            # 缓存数据库类型: no/bigcache/freecache/redis
        # bigcache 配置 (type=bigcache)
        bigCache:
          shards: 1024
          cleanTimeSec: 60
          maxEntriesInWindow: 0
          maxEntrySize: 0
          hardMaxCacheSize: 0       # MB, 0=无限制
          exactExpire: false
        # freecache 配置 (type=freecache)
        freeCache:
          sizeMB: 256
        # redis 配置 (type=redis)
        redisName: "default"        # 使用已有 redis 组件
        # 或独立 redis 配置:
        redis:
          address: "localhost:6379"
          # ... (同 redis 组件配置)
```

## 使用方式

### 1. 获取缓存实例

```go
import "github.com/zly-app/cache/v2"

// 获取默认缓存（通过 Creator）
c := cache.GetCacheCreator().GetDefCache()

// 获取命名缓存
c := cache.GetCacheCreator().GetCache("my-cache")
```

### 2. Get — 读取（缓存未命中自动加载）

```go
var result MyData
err := cache.GetCacheCreator().GetDefCache().Get(ctx, "user:123", &result,
    cache.WithLoadFn(func(ctx context.Context, key string) (interface{}, error) {
        // key 为缓存键名，可用于日志或条件加载
        data, err := loadFromDB(ctx, 123)
        return data, err
    }),
)
```

### 3. Set — 写入

```go
err := cache.GetCacheCreator().GetDefCache().Set(ctx, "user:123", &data,
    cache.WithExpire(600),  // 可选: 覆盖默认过期时间
)
```

### 4. Del — 删除

```go
err := cache.GetCacheCreator().GetDefCache().Del(ctx, "user:123")              // 单个 key
err := cache.GetCacheCreator().GetDefCache().Del(ctx, "user:123", "user:456")  // 多个 key（可变参数）
```

### 5. SingleFlightDo — 单飞执行（忽略缓存）

```go
var result MyData
err := cache.GetCacheCreator().GetDefCache().SingleFlightDo(ctx, "user:123", &result,
    cache.WithLoadFn(func(ctx context.Context, key string) (interface{}, error) {
        return loadFromDB(ctx, 123)
    }),
)
```

### 6. 选项

```go
cache.WithLoadFn(fn)                          // 设置加载函数，签名为 func(ctx context.Context, key string) (interface{}, error)
cache.WithExpire(seconds)                     // 覆盖过期时间（秒，<0 表示永不过期，0 使用默认值）
cache.WithForceLoad(dontWriteCache bool)      // 强制从加载函数获取，dontWriteCache=true 时不写入缓存
cache.WithSerializer(serializer)              // 覆盖序列化器（core.ISerializer 接口）
cache.WithCompactor(compactor)                // 覆盖压缩器（core.ICompactor 接口）
```

### 典型使用场景: 查询缓存

```go
func (q *QueryEngine) GetUserInfo(ctx context.Context, userID int64) (*UserInfo, error) {
    key := fmt.Sprintf("user:info:%d", userID)
    var info UserInfo
    err := cache.GetCacheCreator().GetDefCache().Get(ctx, key, &info,
        cache.WithLoadFn(func(ctx context.Context, key string) (interface{}, error) {
            return q.loadUserInfoFromDB(ctx, userID)
        }),
        cache.WithExpire(300),
    )
    return &info, err
}
```
