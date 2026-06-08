package core

import "fmt"

// ValueType 表示 5 种 RESP 数据类型之一。
// 字节值对应 RESP 前缀字符：+ - : $ *
type ValueType byte

const (
	TypeString  = '+'
	TypeError   = '-'
	TypeInteger = ':'
	TypeBulk    = '$'
	TypeArray   = '*'
)

// Value 是所有 RESP 值类型必须实现的接口。
// 任何可以通过 RESP 协议发送的值都要实现此接口。
type Value interface {
	// Type 返回该值的类型（如 TypeString、TypeError 等）
	Type() ValueType
	// String 返回人类可读的调试表示
	String() string
}

// --- 简单字符串 (+) ---

// StringVal 表示 RESP 简单字符串。
// 简单字符串简短、人类可读，不能包含 \r 或 \n。
type StringVal struct {
	Str string
}

func (s *StringVal) Type() ValueType { return TypeString }
func (s *StringVal) String() string  { return fmt.Sprintf("StringVal(%q)", s.Str) }

func NewStringVal(s string) *StringVal {
	return &StringVal{Str: s}
}

// --- 错误 (-) ---

// ErrorVal 表示 RESP 错误。
// 错误包含前缀（如 "ERR"、"WRONGTYPE"）和消息内容。
type ErrorVal struct {
	Msg string
}

func (e *ErrorVal) Type() ValueType { return TypeError }
func (e *ErrorVal) String() string  { return fmt.Sprintf("ErrorVal(%q)", e.Msg) }

func NewErrorVal(msg string) *ErrorVal {
	return &ErrorVal{Msg: msg}
}

// --- 整数 (:) ---

// IntVal 表示 RESP 整数（64 位有符号）。
type IntVal struct {
	N int64
}

func (i *IntVal) Type() ValueType { return TypeInteger }
func (i *IntVal) String() string  { return fmt.Sprintf("IntVal(%d)", i.N) }

func NewIntVal(n int64) *IntVal {
	return &IntVal{N: n}
}

// --- 批量字符串 ($) ---

// BulkVal 表示 RESP 批量字符串（二进制安全）。
// Data 为 nil 表示"null 批量字符串"（就像 GET 不存在的 key）。
// Data 为非 nil（可能为空切片）表示普通批量字符串。
type BulkVal struct {
	Data []byte
}

func (b *BulkVal) Type() ValueType { return TypeBulk }
func (b *BulkVal) String() string {
	if b.Data == nil {
		return "BulkVal(null)"
	}
	return fmt.Sprintf("BulkVal(%q)", string(b.Data))
}

func NewBulkVal(data []byte) *BulkVal {
	return &BulkVal{Data: data}
}

func NewNullBulkVal() *BulkVal {
	return &BulkVal{Data: nil}
}

// --- 数组 (*) ---

// ArrayVal 表示 RESP 数组（有序的 Value 集合）。
// Items 为 nil 表示"null 数组"。
// Items 为非 nil（可能为空切片）表示普通数组。
type ArrayVal struct {
	Items []Value
}

func (a *ArrayVal) Type() ValueType { return TypeArray }
func (a *ArrayVal) String() string {
	if a.Items == nil {
		return "ArrayVal(null)"
	}
	return fmt.Sprintf("ArrayVal(len=%d)", len(a.Items))
}

func NewArrayVal(items []Value) *ArrayVal {
	return &ArrayVal{Items: items}
}

func NewNullArrayVal() *ArrayVal {
	return &ArrayVal{Items: nil}
}

var (
	_ Value = (*StringVal)(nil)
	_ Value = (*ErrorVal)(nil)
	_ Value = (*IntVal)(nil)
	_ Value = (*BulkVal)(nil)
	_ Value = (*ArrayVal)(nil)
)
