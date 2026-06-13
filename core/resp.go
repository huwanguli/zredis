package core

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Reader 从 io.Reader 中读取并解析 RESP 协议数据。
type Reader struct {
	r *bufio.Reader
}

// NewReader 创建一个 RESP Reader。
func NewReader(rd io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(rd)}
}

// ReadValue 读取并解析一个完整的 RESP 值。
func (r *Reader) ReadValue() (Value, error) {
	prefix, err := r.r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch prefix {
	case '+':
		return r.readSimpleString()
	case '-':
		return r.readError()
	case ':':
		return r.readInteger()
	case '$':
		return r.readBulkString()
	case '*':
		return r.readArray()
	default:
		return nil, fmt.Errorf("invalid RESP prefix: %c", prefix)
	}
}

// readLine 读到 \r\n 为止，返回不含 \r\n 的字节。
func (r *Reader) readLine() ([]byte, error) {
	body := make([]byte, 0)
	for {
		line, isPrefix, err := r.r.ReadLine()
		if err != nil {
			return nil, err
		}
		body = append(body, line...)
		if !isPrefix {
			break
		}
	}
	return body, nil
}

// readSimpleString 读简单字符串 (+)，返回 StringVal。
func (r *Reader) readSimpleString() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	return NewStringVal(string(line)), nil
}

// readError 读错误 (-)，返回 ErrorVal。
func (r *Reader) readError() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	return NewErrorVal(string(line)), nil
}

// readInteger 读整数 (:)，返回 IntVal。
func (r *Reader) readInteger() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	n, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer: %s", string(line))
	}
	return NewIntVal(n), nil
}

// readBulkString 读批量字符串 ($)，返回 BulkVal。
func (r *Reader) readBulkString() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	length, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bulk string length: %s", line)
	}
	if length == -1 {
		return NewNullBulkVal(), nil
	}
	data := make([]byte, length+2)
	if _, err := io.ReadFull(r.r, data); err != nil {
		return nil, err
	}
	return NewBulkVal(data[:length]), nil
}

// readArray 读数组 (*)，返回 ArrayVal。
func (r *Reader) readArray() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	count, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %s", line)
	}
	if count == -1 {
		return NewNullArrayVal(), nil
	}
	items := make([]Value, 0, count)
	for i := int64(0); i < count; i++ {
		item, err := r.ReadValue()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return NewArrayVal(items), nil
}

// Writer 将 RESP Value 编码为字节流写入 io.Writer。
type Writer struct {
	w *bufio.Writer
}

// NewWriter 创建一个 RESP Writer。
func NewWriter(wr io.Writer) *Writer {
	return &Writer{w: bufio.NewWriter(wr)}
}

// WriteValue 将一个 Value 按 RESP 格式编码并写入。
func (w *Writer) WriteValue(v Value) error {
	// TODO: type switch on v
	//   *StringVal → 写 +str\r\n
	//   *ErrorVal   → 写 -msg\r\n
	//   *IntVal     → 写 :n\r\n
	//   *BulkVal    → 写 $len\r\ndata\r\n（Data == nil → $-1\r\n）
	//   *ArrayVal   → 写 *n\r\n + 递归（Items == nil → *-1\r\n）
	return nil
}

// Flush 刷新缓冲区。
func (w *Writer) Flush() error {
	return w.w.Flush()
}

var _ = fmt.Sprintf
