package core

import (
	"io"
	"os"
	"sync"
)

// writeCommands 所有需要记录到 AOF 的写命令。
var writeCommands = map[string]bool{
	// 通用
	"DEL": true, "FLUSHALL": true, "EXPIRE": true, "PERSIST": true,
	// String
	"SET": true, "GETSET": true, "SETEX": true, "SETNX": true, "MSET": true,
	"INCR": true, "INCRBY": true, "DECR": true, "DECRBY": true, "APPEND": true,
	// Hash
	"HSET": true, "HDEL": true,
	// List
	"LPUSH": true, "RPUSH": true, "LPOP": true, "RPOP": true, "LSET": true,
	// Set
	"SADD": true, "SREM": true,
	// ZSet
	"ZADD": true, "ZREM": true,
}

// AOF 管理 Append-Only File 的写入和重放。
type AOF struct {
	file *os.File
	mu   sync.Mutex
	w    *Writer
}

// NewAOF 打开或创建 AOF 文件，返回 AOF 实例。
func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	w := NewWriter(f)
	return &AOF{
		file: f,
		w:    w,
	}, nil
}

// Write 将一条命令编码写入 AOF 文件。
func (a *AOF) Write(cmd *ArrayVal) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.w.WriteValue(cmd); err != nil {
		return err
	}
	if err := a.w.Flush(); err != nil {
		return err
	}
	return a.file.Sync()
}

// Load 读取 AOF 文件，逐条重放到 Store 中恢复数据。
func (a *AOF) Load(s *Store, d *Dispatcher) error {
	a.file.Seek(0, 0)
	reader := NewReader(a.file)
	for {
		val, err := reader.ReadValue()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if cmd, ok := val.(*ArrayVal); ok {
			d.Dispatch(s, cmd)
		}
	}
	return nil
}

// Close 关闭 AOF 文件。
func (a *AOF) Close() error {
	return a.file.Close()
}

// IsWriteCmd 判断是否为需要记录的写命令。
func IsWriteCmd(name string) bool {
	return writeCommands[name]
}

var _ = os.O_APPEND
var _ = sync.Mutex{}
