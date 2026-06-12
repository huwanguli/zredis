# Step 7: Set 类型

## 这是什么？

Redis Set 是一个无序的、不重复的字符串集合。底层用哈希表实现，支持交并差运算。

```
SADD tags "go" "redis"   → 2
SISMEMBER tags "go"      → true
SINTER set1 set2         → 交集
```

## 实现方式

`map[string]struct{}` — Go 惯用的集合表示。增删查皆为 O(1)。

## 架构模式

```
set.go
  内部方法：getSet / lookupSet / sadd / srem   ← 不锁
  导出方法：SAdd / SRem / SMembers ...         ← Lock + 调内部
```

## 要实现的命令

| 命令 | 语义 | 返回值 |
|------|------|--------|
| `SADD key member [member ...]` | 添加成员 | 新增数量 |
| `SREM key member [member ...]` | 删除成员 | 删除数量 |
| `SMEMBERS key` | 所有成员 | []string |
| `SISMEMBER key member` | 判断存在 | bool |
| `SCARD key` | 成员数量 | int |
| `SINTER key1 key2` | 交集 | []string |
| `SUNION key1 key2` | 并集 | []string |
| `SDIFF key1 key2` | 差集 (key1 - key2) | []string |
