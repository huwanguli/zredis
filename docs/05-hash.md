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

在 Store 里加一个 `hashes` 字段，专门存 Hash 数据：

```go
type Store struct {
    ...
    hashes map[string]map[string]string  // key → (field → value)
}
```

这是最简方案，后续可以统一重构为 `map[string]Value` 接口。

## 要实现的命令

| 命令 | 语义 | 返回值 |
|------|------|--------|
| `HSET key field value [field value ...]` | 设值 | 新增 field 数量 |
| `HGET key field` | 取值 | 值和是否存在 |
| `HDEL key field [field ...]` | 删 field | 删除数量 |
| `HGETALL key` | 取全部 | field→value 的 map |
| `HEXISTS key field` | 判断存在 | bool |
| `HLEN key` | field 数量 | int |
| `HKEYS key` | 所有 field 名 | []string |
| `HVALS key` | 所有 value | []string |

## 细节

- key 不存在时，除 HGET/HEXISTS 返回 false 外，其余命令返回 0 / 空
- HSET 支持多对 field-value（变长参数），返回**新增**的数量（不是修改的）
- 继续用内部/导出拆分模式，HSET 内部版供后续组合命令复用
