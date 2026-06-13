# Step 10: RESP 编码器

## 这是什么？

编码器把 Step 1 的 `Value` 转回 RESP 字节格式，发给客户端。

## 编码规则

| 类型 | Value 判断 | RESP 输出 |
|------|-----------|----------|
| StringVal | `v.(*StringVal)` | `+str\r\n` |
| ErrorVal | `v.(*ErrorVal)` | `-msg\r\n` |
| IntVal | `v.(*IntVal)` | `:n\r\n` |
| BulkVal | `Data == nil` | `$-1\r\n` |
| BulkVal | `Data != nil` | `$len\r\ndata\r\n` |
| ArrayVal | `Items == nil` | `*-1\r\n` |
| ArrayVal | `Items != nil` | `*n\r\n` + 递归 |

## 实现方案

```go
type Writer struct {
    w *bufio.Writer
}

func (w *Writer) WriteValue(v Value) error { ... }
```

`WriteValue` 用 type switch 分发到各类型的写方法。

提示：`fmt.Fprintf` 可以直接写格式化字符串。BulkVal 先写 `$len\r\n`，再写 `data\r\n`（两次 write）。Array 先写 `*n\r\n`，再遍历递归。

## 你要写的

`resp.go` 中 `Writer` 的编码方法，约 40 行。
