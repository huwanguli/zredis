# Step 8: Sorted Set + 跳表

## 什么是跳表？

普通链表查找需要 O(n)。跳表在链表上加多层"索引"，高层跳得远，低层走得细，实现 O(log n) 查找。

```
Level 2:  HEAD ─────────────── 30:c ────→ NULL
Level 1:  HEAD ──── 10:a ───── 30:c ────→ NULL
Level 0:  HEAD → 10:a → 20:b → 30:c → NULL

真实数据只有 level 0，上面的 layer 是索引。
```

## 数据结构

```go
type SkipListLevel struct {
    Forward *SkipListNode  // 这一层指向的下一个节点
    Span    int64          // 从当前节点到 Forward 跳过了几个元素（含终点不含起点）
}

type SkipListNode struct {
    Member   string
    Score    float64
    Backward *SkipListNode    // 后退指针（只在 level 0 有用）
    Levels   []SkipListLevel  // 各层的前向指针和跨度
}

type SkipList struct {
    Header *SkipListNode  // 哨兵节点，不存数据
    Tail   *SkipListNode
    Length int64
    Level  int            // 当前最高层级（1-based）
}
```

**rank 的定义**：一个节点的 rank = 排在它前面的节点数量（0-based）。Header 不算在内。

**span 的定义**：从节点 A 到节点 B 的 span = rank(B) - rank(A)，即中间经过了多少个节点（包含 B，不包含 A）。

## 插入算法（Insert）

以插入 `20:b` 到已有列表 `10:a → 30:c` 为例（Length=2，Level=2）：

```
当前状态：
Level 1: HEAD --span=2--> 30:c    (rank(30:c)=1, span=2=1-(-1)+1-1=2? 实际上 rank(30:c)=1, rank(HEAD)按-1算, span = 1-(-1) = 2)
Level 0: HEAD --span=1--> 10:a --span=1--> 30:c
```

### 步骤 1：搜索插入位置

从最高层往下，每层记录 update[i]（该层的前驱节点）和 rank[i]（到达该位置的累计排名）。

```go
update[skipListMaxLevel]  // 每层的前驱
rank[skipListMaxLevel]    // 每层已跨越的节点数

x = Header
for i = Level-1 down to 0:
    rank[i] = (i == Level-1) ? 0 : rank[i+1]  // 继承上层排名
    // 前进条件：forward 的 (score, member) < 目标的 (score, member)
    while x.Levels[i].Forward != nil &&
          (x.Levels[i].Forward.Score < score ||
           (x.Levels[i].Forward.Score == score &&
            x.Levels[i].Forward.Member < member)):
        rank[i] += x.Levels[i].Span
        x = x.Levels[i].Forward
    update[i] = x
```

对本例（score=20, member="b"）：

```
i=1: HEAD.Levels[1].Forward=30:c, 30>20 → 不前进
     update[1]=HEAD, rank[1]=0

i=0: HEAD.Levels[0].Forward=10:a, 10<20 → 前进
     rank[0] += 1 = 1, x = 10:a
     10:a.Levels[0].Forward=30:c, 30>20 → 不前进
     update[0]=10:a, rank[0]=1
```

结果：`update[0]=10:a, rank[0]=1`, `update[1]=HEAD, rank[1]=0`

### 步骤 2：检查 member 是否已存在

```go
x = update[0].Levels[0].Forward
if x != nil && x.Score == score && x.Member == member {
    x.Score = score  // 存在则只更新分数
    return nil
}
```

### 步骤 3：随机生成层数

```go
level := 1
for rand.Float64() < 0.25 && level < 32 {
    level++
}
```

### 步骤 4：处理新层数高于当前 max level 的情况

```go
if level > sl.Level {
    for i = sl.Level; i < level; i++ {
        rank[i] = 0
        update[i] = sl.Header
        update[i].Levels[i].Span = sl.Length
    }
    sl.Level = level
}
```

### 步骤 5：在各层插入新节点

对 i = 0 到 level-1：

```go
// 挂链
newNode.Levels[i].Forward = update[i].Levels[i].Forward
update[i].Levels[i].Forward = newNode

// 计算 span
// 距离 d = rank[0] - rank[i]，即 update[i] 到插入位置之间有多少个节点
newNode.Levels[i].Span = update[i].Levels[i].Span - d
update[i].Levels[i].Span = d + 1
```

以本例 newLevel=2 代入验证：

```
i=0: d = rank[0] - rank[0] = 0
     update[0]=10:a, 原 span=1
     newNode.Span = 1 - 0 = 1    (20:b → 30:c 经过 1 个节点：30:c 自己) ✓
     update[0].Span = 0 + 1 = 1  (10:a → 20:b 经过 1 个节点：20:b 自己) ✓

i=1: d = rank[0] - rank[1] = 1 - 0 = 1
     update[1]=HEAD, 原 span=2
     newNode.Span = 2 - 1 = 1    (20:b → 30:c) ✓
     update[1].Span = 1 + 1 = 2  (HEAD → 20:b, 经过 10:a 和 20:b) ✓
```

### 步骤 6：高于新节点层数的层，span 全部 +1

```go
for i = level; i < sl.Level; i++ {
    update[i].Levels[i].Span++
}
```

因为多了一个新节点，高层跳过时需要多跨一步。

### 步骤 7：设置 backward 指针和 Tail

```go
newNode.Backward = nil
if update[0] != sl.Header {
    newNode.Backward = update[0]
}
if newNode.Levels[0].Forward != nil {
    newNode.Levels[0].Forward.Backward = newNode
} else {
    sl.Tail = newNode
}
sl.Length++
```

## 按排名查找（GetElementByRank）

反过来用 span 定位：

```go
func (sl *SkipList) GetElementByRank(rank int64) *SkipListNode {
    if rank < 0 || rank >= sl.Length { return nil }
    traversed := int64(0)  // 已跨越的节点数
    x := sl.Header
    for i := sl.Level - 1; i >= 0; i-- {
        for x.Levels[i].Forward != nil && traversed + x.Levels[i].Span <= rank {
            traversed += x.Levels[i].Span
            x = x.Levels[i].Forward
        }
    }
    return x.Levels[0].Forward
}
```

## 获取排名（GetRank）

先找到节点，再累计 span：

```go
func (sl *SkipList) GetRank(score float64, member string) int64 {
    rank := int64(0)
    x := sl.Header
    for i := sl.Level - 1; i >= 0; i-- {
        for x.Levels[i].Forward != nil &&
            (x.Levels[i].Forward.Score < score ||
             (x.Levels[i].Forward.Score == score &&
              x.Levels[i].Forward.Member <= member)) {
            rank += x.Levels[i].Span
            x = x.Levels[i].Forward
        }
    }
    if x.Member == member { return rank }
    return -1  // 未找到
}
```

注意比较条件用了 `<=`：排名计算时要包含 member 自己，所以 `Member <= member`。

## 删除（Delete）

类似 Insert，找到各层前驱，更新指针和 span：

```go
func (sl *SkipList) Delete(score float64, member string) bool {
    update[skipListMaxLevel]
    x := sl.Header
    for i := sl.Level - 1; i >= 0; i-- {
        for x.Levels[i].Forward != nil && (forward < target) {
            x = x.Levels[i].Forward
        }
        update[i] = x
    }
    x = x.Levels[0].Forward  // 目标节点
    if x == nil || x.Score != score || x.Member != member { return false }

    for i := 0; i < sl.Level; i++ {
        if update[i].Levels[i].Forward == x {
            // 将 update 的 span 合并到跨过被删节点
            update[i].Levels[i].Span += x.Levels[i].Span - 1
            update[i].Levels[i].Forward = x.Levels[i].Forward
        } else {
            update[i].Levels[i].Span--
        }
    }

    // 更新 backward
    if x.Levels[0].Forward != nil {
        x.Levels[0].Forward.Backward = x.Backward
    } else {
        sl.Tail = x.Backward
    }

    // 降低 level（如果最高层空了）
    for sl.Level > 1 && sl.Header.Levels[sl.Level-1].Forward == nil {
        sl.Level--
    }
    sl.Length--
    return true
}
```

## Find

简单的搜索，不累计 rank：

```go
func (sl *SkipList) Find(score float64, member string) *SkipListNode {
    x := sl.Header
    for i := sl.Level - 1; i >= 0; i-- {
        for x.Levels[i].Forward != nil &&
            (x.Levels[i].Forward.Score < score ||
             (x.Levels[i].Forward.Score == score &&
              x.Levels[i].Forward.Member < member)) {
            x = x.Levels[i].Forward
        }
    }
    x = x.Levels[0].Forward
    if x != nil && x.Score == score && x.Member == member {
        return x
    }
    return nil
}
```

## GetRange

```go
func (sl *SkipList) GetRange(start, stop int64) []string {
    if start < 0 { start = sl.Length + start }
    if stop < 0 { stop = sl.Length + stop }
    if start < 0 { start = 0 }
    if stop >= sl.Length { stop = sl.Length - 1 }
    if start > stop { return nil }

    node := sl.GetElementByRank(start)
    result := make([]string, 0, stop-start+1)
    for i := start; i <= stop && node != nil; i++ {
        result = append(result, node.Member)
        node = node.Levels[0].Forward
    }
    return result
}
```

## 你要写的文件

**`core/skiplist.go`** — 7 个方法：
1. `NewSkipList` — 创建 header（Levels 长度 32）
2. `Find` — 按上述代码
3. `Insert` — 按上述算法，最核心
4. `GetElementByRank` — 按上述代码
5. `GetRank` — 按上述代码
6. `GetRange` — 调 GetElementByRank 即可
7. `Delete` — 按上述算法

**`core/zset.go`** — 基于 skiplist 的薄壳，ZAdd/ZRange/ZRank/ZScore/ZRem/ZCard。
