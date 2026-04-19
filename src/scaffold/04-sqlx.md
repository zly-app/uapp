# SQLx — MySQL/PostgreSQL 等 SQL 数据库

> 导入: `github.com/zly-app/component/sqlx`

## 功能概览

- 支持 5 种数据库驱动: MySQL, PostgreSQL, MSSQL, ClickHouse, SQLite3
- 基于 `jmoiron/sqlx` 封装，提供更简洁的 API
- 自动集成 trace/metrics/日志 (通过 Filter 链)
- 支持事务（标准/扩展），事务隔离级别和只读设置
- 支持多实例（按名获取）

## 配置 (`components.sqlx.{name}`)

```yaml
components:
  sqlx:
    default:                            # 组件名
      driver: "mysql"                   # 驱动: mysql/postgres/mssql/clickhouse/sqlite3
      source: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=true"
      maxIdleConns: 3                   # 最大空闲连接
      maxOpenConns: 10                  # 最大连接数
      connMaxLifetime: 0                # 最大续航时间(毫秒, 0=无限)
```

## 使用方式

### 1. 获取客户端

```go
import "github.com/zly-app/component/sqlx"

// 获取默认客户端（包级便捷函数）
client := sqlx.GetDefClient()

// 获取命名客户端（包级便捷函数）
client := sqlx.GetClient("my-db")

// 通过 Creator 接口获取（等价于上面的包级函数）
client := sqlx.GetCreator().GetDefClient()
client := sqlx.GetCreator().GetClient("my-db")
```

> **提示**: `sqlx.GetDefClient()` / `sqlx.GetClient(name)` 是包级便捷函数，可直接调用，无需先获取 Creator。

### 2. 查询

```go
// 多行查询
var users []User
err := client.Find(ctx, &users, "SELECT * FROM users WHERE age > ?", 18)

// 单行查询 (未找到返回 sqlx.ErrNoRows)
var user User
err := client.FindOne(ctx, &user, "SELECT * FROM users WHERE id = ?", id)

// 列查询
var names []interface{}
err := client.FindColumn(ctx, &names, "SELECT name FROM users")
```

### 3. 执行

```go
result, err := client.Exec(ctx, "INSERT INTO users (name, age) VALUES (?, ?)", "Alice", 20)
```

### 4. 逐行处理

```go
err := client.Query(ctx, func(rows *sql.Rows) error {
    var name string
    var age int
    if err := rows.Scan(&name, &age); err != nil {
        return err
    }
    // 处理每一行...
    return nil // 返回 sqlx.ErrBreakNext 可提前停止
}, "SELECT name, age FROM users")
```

### 5. 事务

```go
// 标准事务（回调函数接收 ctx）
// Tx 接口方法: Exec, Query, FindColumn, FindToStructs
err := sqlx.GetCreator().GetDefClient().Transaction(ctx, func(ctx context.Context, tx sqlx.Tx) error {
    _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromID)
    if err != nil { return err }
    _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toID)
    return err
})

// 扩展事务（支持 struct 扫描，回调函数接收 ctx）
// Txx 接口方法: Find, FindOne, FindColumn, Exec, Query + Unsafe 模式
err := sqlx.GetCreator().GetDefClient().TransactionX(ctx, func(ctx context.Context, tx sqlx.Txx) error {
    var user User
    err := tx.FindOne(ctx, &user, "SELECT * FROM users WHERE id = ? FOR UPDATE", id)
    // ...
    return nil
})

// 事务选项
err := sqlx.GetCreator().GetDefClient().Transaction(ctx, fn,
    sqlx.WithTxIsolation(sql.LevelSerializable),
    sqlx.WithTxReadOnly(true),
)
```

> **注意**: `sqlx.Tx` 接口没有 `FindOne` 方法（仅有 `Exec`、`Query`、`FindColumn`、`FindToStructs`）；如需在事务中使用 `FindOne`，请使用 `TransactionX` 获取 `sqlx.Txx` 接口。

### 6. Unsafe 模式

```go
// 获取 Unsafe 客户端（允许将未映射的列忽略，而非报错）
unsafeClient := client.Unsafe()
err := unsafeClient.FindOne(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
```

### 7. 动态 SQL 构建

推荐使用 `github.com/didi/gendry/builder`:

```go
import "github.com/didi/gendry/builder"

where := map[string]interface{}{
    "id": id,
}
selectField := []string{"id", "name", "age"}
cond, vals, err := builder.BuildSelect("users", where, selectField)

var user User
err = db.GetSqlx().FindOne(ctx, &user, cond, vals...)
```

### 7. 获取底层 DB

```go
db := sqlx.GetCreator().GetDefClient().GetDB()  // *sqlx.DB
```
