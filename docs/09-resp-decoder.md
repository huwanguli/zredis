# Step 9: RESP 解码器

## 这是什么？

RESP（Redis Serialization Protocol）是 Redis 客户端和服务器通信的二进制协议。
解码器把字节流解析成 Step 1 定义的 `Value` 类型。

## 格式

| 前缀 | 类型 | 编码 | 示例 |
|------|------|------|------|
| `+` | 简单字符串 | `+内容\r\n` | `+OK\r\n` |
| `-` | 错误 | `-消息\r\n` | `-ERR unknown\r\n` |
| `:` | 整数 | `:数字\r\n` | `:1000\r\n` |
| `$` | 批量字符串 | `$长度\r\n数据\r\n` | `$5\r\nhello\r\n` |
| `*` | 数组 | `*数量\r\n元素1元素2...` | `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n` |

特殊值：`$-1\r\n` = null 批量字符串，`*-1\r\n` = null 数组。

## 实现方案

```go
type Reader struct {
    r *bufio.Reader
}

func (r *Reader) ReadValue() (Value, error) { ... }
```

`ReadValue()` 读第一个字节判断类型，然后按格式解析：
- `+` → 读到 `\r\n` 为止
- `-` → 读到 `\r\n` 为止
- `:` → 读到 `\r\n`，`strconv.ParseInt`
- `$` → 读长度，-1 返回 NullBulkVal；否则读指定长度 + `\r\n`
- `*` → 读数量，-1 返回 NullArrayVal；否则递归调 ReadValue 读每个元素

## 你要写的

`resp.go` 中 `Reader` 的 `ReadValue()` 方法，约 60 行。
