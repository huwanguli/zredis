# Step 3: TTL 与过期清理

## 这是什么？

Redis 的每个 key 都可以设置过期时间（TTL，Time To Live）。到期后 key 自动删除。
比如 `SET session abc EX 3600` 表示 3600 秒后这个 key 消失。

这一步你要给 Store 添加过期机制。

## 为什么重要？

没有 TTL，Redis 就只是一个带锁的 map。TTL 是 Redis 作为缓存的核心能力——
自动淘汰过期数据，防止内存无限增长。

## 两种过期策略

### 1. 惰性删除（Lazy Expiration）

每次访问 key 时检查是否过期。如果过期了，当场删除，返回"不存在"。

**优点**：简单，不额外消耗 CPU
**缺点**：不访问的过期 key 永远不会被删除，内存泄漏

### 2. 定期删除（Periodic Expiration）

后台 goroutine 每隔一段时间随机抽一部分 key 检查，过期就删。

**优点**：不会内存泄漏
**缺点**：额外的 CPU 开销

**Redis 的做法：两者都用。** 惰性删除保证访问时不会拿到过期数据，
定期删除保证"冷"过期 key 最终被清理。

## 核心问题：Get 的锁升级

引入 TTL 后，`Get` 不能再只用 `RLock` —— 因为发现过期 key 时需要**删除**它（写操作）。

你有两个选择：
- **直接换 `Lock`**：简单，但牺牲了并发读的性能
- **先读再升级**：RLock 读到过期 key → 解锁 → Lock → 二次检查 → 删除

这里建议用 `Lock`，简单可靠。真正的 Redis 是单线程事件循环，不存在这种取舍。

## 真实世界的参考

Redis 源码中，过期机制在 `expire.c` 和 `db.c`：
- `expireIfNeeded()` 在每次访问 key 时调用（惰性删除）
- `activeExpireCycle()` 在 `serverCron()` 中周期性调用，每次扫描一部分随机 key
- 过期时间存储在 `redisDb->expires` 字典中（key → 毫秒级 Unix 时间戳）

## 你要构建的

**修改 `store.go`：**
- Store 结构体增加 `expires map[string]time.Time` 字段
- `NewStore` 初始化 expires
- `Get` 换 `Lock`，加惰性过期检查
- `Del` 同时清理 expires
- `Flush` 同时清理 expires

**新增 `ttl.go`：**
- `Expire(key, ttl)` — 设置过期时间
- `TTL(key)` — 查询剩余时间
- `Persist(key)` — 移除过期时间
- `CleanExpired(batchSize)` — 定期清理，随机抽检 batchSize 个 key
