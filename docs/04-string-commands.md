# Step 4: String 命令

## 这是什么？

在 Redis 中，"String" 是最基础的数据类型，一个 key 对应一个字符串值。
这一步骤你要实现 String 类型的全部操作命令，在 Store 的基础上封装出 Redis 语义。

## 你要实现的命令

| 命令 | 语义 | 返回值 |
|------|------|--------|
| `INCR key` | 值+1 | 新值 |
| `INCRBY key delta` | 值+delta | 新值 |
| `DECR key` | 值-1 | 新值 |
| `DECRBY key delta` | 值-delta | 新值 |
| `APPEND key value` | 追加 | 新长度 |
| `STRLEN key` | 字符串长度 | 长度 |
| `MGET key1 key2 ...` | 批量获取 | null/值的数组 |
| `MSET k1 v1 k2 v2` | 批量设置 | 无 |
| `GETSET key value` | 设新值返回旧值 | 旧值 |
| `SETEX key ttl value` | 设值+过期 | 无 |
| `SETNX key value` | 不存在才设 | 是否设置成功 |

## 为什么放在 Store 上？

这些命令本质都是对 Store 的原子操作。例如 `INCR` 需要"读 → 解析 → +1 → 写回"，
整个过程必须在**同一把锁内**完成，否则并发 INCR 会丢数据。

## INCR/DECR 的细节

- key 不存在时，**先用 0 初始化**，然后执行加减
- key 存在但值不是数字时，返回错误
- 用 `strconv.ParseInt` 解析，`strconv.FormatInt` 写回

## MGET 的返回值

返回 `[]*BulkVal`，用到了 Step 1 的值类型：
- key 存在 → `NewBulkVal([]byte(value))`
- key 不存在 → `NewNullBulkVal()`

## 真实世界的参考

Redis 的 `t_string.c` 实现了所有 String 命令。核心技巧：
- `getGenericCommand()` 复用于 GET/GETSET
- `incrDecrCommand()` 复用于 INCR/INCRBY/DECR/DECRBY
- 所有操作都在单线程内完成，天然原子性（我们靠锁保证）

## 你要构建的

在 `core/string_cmd.go` 中为 `*Store` 添加 11 个方法。DECR 和 DECRBY 可以复用 IncrBy 逻辑。
