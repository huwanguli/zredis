# Step 12: TCP Server

## 这是什么？

把前面的组件串起来：监听端口 → 接受连接 → 每个连接循环读 RESP 命令 → 分发执行 → 写回结果。

## 设计

```go
type Server struct {
    store *Store
    disp  *Dispatcher
}

func (srv *Server) Listen(addr string) error
```

- `Listen` 调 `net.Listen`，主循环 `Accept`
- 每个连接起一个 goroutine 执行 `handleConn`
- `handleConn` 用 `NewReader(conn)` 读命令，`NewWriter(conn)` 写结果
- 读到 EOF 或错误时关闭连接

## 流程

```
conn → Reader.ReadValue() → *ArrayVal → Dispatcher.Dispatch() → Value → Writer.WriteValue() → conn
```

## 你要写的

`server.go`，实现 `Server` 的 `Listen` 和 `handleConn` 方法。约 40 行。

提示：
- `net.Listen("tcp", addr)` 返回 `net.Listener`
- `listener.Accept()` 返回 `net.Conn, error`
- `io.EOF` 表示客户端断开，正常退出
