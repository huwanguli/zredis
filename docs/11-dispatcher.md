# Step 11: 命令分发器

## 这是什么？

RESP 解码后得到一个 `*ArrayVal`，例如 `["SET", "mykey", "hello"]`。
分发器根据第一个元素（命令名）找到对应的 Store 方法，把剩下的元素转成参数调用，返回 `Value`。

## 流程

```
客户端字节流 → Reader.ReadValue() → *ArrayVal → Dispatch() → Value → Writer.WriteValue() → 字节流
```

## 设计

```go
type CommandFunc func(s *Store, args [][]byte) Value

type Dispatcher struct {
    cmds map[string]CommandFunc
}
```

用一个 `map[string]CommandFunc` 注册所有命令。`Dispatch` 方法：
1. 取 `Items[0]` 命令名（转大写）
2. 剩余 `Items[1:]` 转成 `[][]byte`
3. 查 map，调用对应函数，返回 Value
4. 未知命令返回 `-ERR unknown command`

## 参数提取

每个命令自己从 `args [][]byte` 中提取参数：
- `string(args[0])` → string
- `strconv.ParseInt(string(args[0]), 10, 64)` → int64
- `len(args)` → 参数个数校验

## 你要写的

`dispatch.go`，注册下列命令并实现它们的 CommandFunc：
- **PING** → `+PONG`
- **SET** key value → Store.Set(key, string)
- **GET** key → BulkVal 或 Null Bulk
- **DEL** key... → Store.Del(...) → 返回删除数量
- **EXISTS** key... → Store.Exists(...) → 返回存在数量
- **INCR** key → Store.Incr(key) → IntVal 或 error
- 再加 **TTL**、**EXPIRE**、**KEYS** 等，你自己挑 3-4 个

约 100 行。
