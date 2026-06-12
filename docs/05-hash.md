# Step 5: Hash 类型

## 这是什么？

Redis Hash 是一个 key 下存储多个 field → value 对的集合。可以理解为"键值对的键值对"：

```
HSET user:1000 name zhangsan age 25
HGET user:1000 name   → "zhangsan"
HGETALL user:1000     → {name: zhangsan, age: 25}
```

## 为什么需要 Hash？

如果只用 String，存一个用户需要多个 key：`user:1000:name`、`user:1000:age`...
Hash 把这些信息聚在一个 key 下，方便批量操作和原子性。

## 实现方式

Hash 数据存在 `DataEntry.Hash` 字段中（和 String 共享同一个 `data` map）：

```go
type DataEntry struct {
    Type   DataType              // DataHash
    String string                // String 类型用
    Hash   map[string]string     // Hash 类型用 ← 这个
}
```

**和 String 的关系**：同一个 key 不能既是 String 又是 Hash。`getHash(key)` 检查 `Type == DataHash`，
`getString(key)` 检查 `Type == DataString`，互斥。

**架构模式**（和 `string_cmd.go` 一致）：

```
hash.go
  内部方法：getHash / hset          ← 不锁，调用者持锁
  导出方法：HSet / HGet / HDel ...  ← Lock + expireIfNeeded + 调内部方法
```

## 要实现的命令

| 命令 | 语义 | 返回值 |
|------|------|--------|
| `HSET key field value [field value ...]` | 设值 | 新增 field 数量 |
| `HGET key field` | 取值 | 值和是否存在 |
| `HDEL key field [field ...]` | 删 field | 删除数量 |
| `HGETALL key` | 取全部 | field→value 的 map（副本） |
| `HEXISTS key field` | 判断存在 | bool |
| `HLEN key` | field 数量 | int |
| `HKEYS key` | 所有 field 名 | []string |
| `HVALS key` | 所有 value | []string |

## 你要写的

只写 **`hset`** 内部方法（`hash.go` 第 20 行）。其余 7 个导出方法我已经写好了，对照理解即可。

`hset` 的逻辑：
1. 调 `s.get(key)` 取 entry
2. 不存在 → 创建 `&DataEntry{Type: DataHash, Hash: make(map[string]string)}`，写入 `s.data[key]`
3. 已存在但 `Type != DataHash` → 创建新 entry 覆盖（Redis 会报 WRONGTYPE，原型简化处理）
4. 遍历 `pairs`（步长 2：field, value），填入 `entry.Hash`
5. 统计**新增**的 field 数量（已存在的修改不计入），返回

参考 `string_cmd.go` 中 `incrBy` 操作 `entry` 的方式。
