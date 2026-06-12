# Step 6: List 类型

## 这是什么？

Redis List 是一个有序的字符串列表，支持从两端推入和弹出。底层用双向链表/快速列表实现，
所有操作都是 O(1)。

```
LPUSH queue "a" "b" "c"    →   c b a  （左端推入）
RPOP queue                   →   a     （右端弹出）
LRANGE queue 0 -1           →   c b   （取全部）
```

## 实现方式

用 Go 标准库 `container/list`（双向链表），存在 `DataEntry.List` 中。
所有头尾操作 O(1)，索引操作需遍历 O(n)。

## 架构模式

和 String/Hash 一致：

```
list.go
  内部方法：getList / lookupList / lpush / rpush   ← 不锁
  导出方法：LPush / RPush / LPop / RPop / LLen ... ← Lock + 调内部方法
```

## 要实现的命令

| 命令 | 语义 | 返回值 |
|------|------|--------|
| `LPUSH key val [val ...]` | 左端推入（可变参数） | 操作后列表长度 |
| `RPUSH key val [val ...]` | 右端推入 | 操作后列表长度 |
| `LPOP key` | 左端弹出 | 值和是否存在 |
| `RPOP key` | 右端弹出 | 值和是否存在 |
| `LLEN key` | 列表长度 | int |
| `LRANGE key start stop` | 截取范围（负索引支持） | []string |
| `LINDEX key index` | 按索引取值（负索引支持） | 值和是否存在 |
| `LSET key index value` | 按索引设值 | error |

## 细节

- key 不存在时 LPUSH/RPUSH 创建新列表；LPOP/RPOP 返回 "", false
- **类型不匹配返回 wrongtype 错误**（和 incrBy/hset 一致）
- **负索引**：-1 表示最后一个，-2 表示倒数第二个。LRANGE/LINDEX 需要转换
- LRANGE 的 stop 可能超出范围，截断即可
- LPOP/RPOP 后如果列表空了，建议保留空列表（或删除 key，二选一）
