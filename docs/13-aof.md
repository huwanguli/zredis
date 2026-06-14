# Step 13: AOF 持久化

## 这是什么？

AOF (Append-Only File) 把所有写命令追加到日志文件，重启时逐条重放，恢复数据。

## 原理

```
客户端 → SET key hello  → Store
                ↘ aof.Write(cmd)   // 追加到 aof.log
```

重启时：
```
aof.log → aof.Load(store) → 逐条 Dispatch → 数据恢复
```

## AOF 文件格式

就是标准的 RESP 命令，一行一条：

```
*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nhello\r\n
*2\r\n$4\r\nINCR\r\n$7\r\ncounter\r\n
```

## 设计

```go
type AOF struct {
    file *os.File
    mu   sync.Mutex
    w    *Writer  // RESP Writer 直接写到文件
}

func NewAOF(path string) (*AOF, error)     // 打开文件
func (a *AOF) Write(cmd *ArrayVal) error   // 编码并写入
func (a *AOF) Load(s *Store) error         // 读取文件，逐条重放
func (a *AOF) Close() error                // 关闭文件
```

## 关键点

1. **Write**: 用 `Writer.WriteValue(cmd)` 写到文件，每次写完要 `Sync()` 保证落盘
2. **Load**: 用 `Reader.ReadValue()` 逐条读出 `*ArrayVal`，调 `Dispatcher.Dispatch` 重放（忽略返回值）
3. **只记录写命令**：维护一个 `writeCommands map[string]bool`，读命令（GET、KEYS 等）不记录
4. **服务端集成**：在 `handleConn` 里，Dispatch 成功后，如果是写命令就 `aof.Write(cmd)`

## 你要写的

`aof.go`，约 70 行。然后改 `server.go` 集成 AOF。
