# Step 2: 并发安全存储

## 这是什么？

Redis 的核心是一个内存中的键值字典 —— 所有的 GET、SET、DEL 等操作本质上都是对这个字典的读写。
这个字典必须支持**并发访问**，因为 Redis 服务端会同时处理多个客户端连接。

在这一步，你要用 Go 的 `sync.RWMutex` + `map` 实现一个线程安全的键值存储引擎。

## 为什么重要？

这是你的 Redis 的数据"心脏"。后面所有命令（无论 String、Hash、List、Set、ZSet）
最终都是在读写这个存储。线程安全是基本要求 —— 一个不安全的 map 在并发读写时会导致
Go runtime 的 fatal error（concurrent map read and map write）。

## 设计决策

### 为什么用 RWMutex 而不是 Mutex？

| 锁类型 | 读-读 | 读-写 | 写-写 | 适合场景 |
|--------|-------|-------|-------|---------|
| `sync.Mutex` | 互斥 | 互斥 | 互斥 | 读写比例差不多 |
| `sync.RWMutex` | **并发** | 互斥 | 互斥 | **读远多于写** |

Redis 的典型场景是读多写少。`RWMutex` 允许多个读操作同时进行，只有写操作才独占锁，
这在高并发读取时性能远好于 `Mutex`。

### 为什么用 map 而不是其他结构？

- `map[string]string` 是 Go 内置的哈希表，查找复杂度 O(1)
- 真正的 Redis 也使用哈希表（`dict.c`，基于渐进式 rehash）
- 对于百万级 key，Go 的 map 完全够用（后续可优化 hash 算法，但现在不需要）

### 为什么存 string 而不是 Value 接口？

当前阶段（Step 2-3）我们只处理字符串类型。后续 Step 4-8 会增加 Hash、List 等类型，
届时扩展存储结构即可。过早抽象会增加复杂度。

## 真实世界的参考

Redis 源码中，键值字典在 `dict.c` / `dict.h`：
- 使用两个哈希表（`ht[2]`）支持渐进式 rehash，避免大字典 rehash 时阻塞
- 每个 dict entry 包含 `key`（sds）、`val`（robj 指针）、`next`（链地址法解决冲突）
- 在我们的 Go 版本中，Go 的 `map` 自带 rehash，大大简化了实现

## 你要构建的

一个 `Store` 结构体，提供：
- `Get(key)` — 读，使用读锁
- `Set(key, value)` — 写，使用写锁
- `Del(key)` — 删，使用写锁
- `Exists(key)` — 判断存在，使用读锁
- `Keys()` — 列出所有 key，使用读锁
- `Len()` — key 数量，使用读锁
- `Flush()` — 清空所有数据，使用写锁
