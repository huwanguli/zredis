package core

import "sync"

// Store 是一个线程安全的内存键值存储引擎。
// 所有对 data 的访问都必须通过锁保护。
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

// TODO: 实现 NewStore() *Store — 创建并返回一个初始化好的 Store（data map 已分配）

// TODO: 实现 Get(key string) (string, bool)
//   使用读锁（RLock/RUnlock），从 data 中取值。
//   如果 key 存在，返回值和 true；否则返回 "" 和 false。

// TODO: 实现 Set(key string, value string)
//   使用写锁（Lock/Unlock），将键值对写入 data。

// TODO: 实现 Del(key string) bool
//   使用写锁。如果 key 存在，删除并返回 true；否则返回 false。

// TODO: 实现 Exists(key string) bool
//   使用读锁。如果 key 存在返回 true，否则 false。

// TODO: 实现 Keys() []string
//   使用读锁。遍历 data，返回所有 key 的切片。

// TODO: 实现 Len() int
//   使用读锁。返回 data 中键值对的数量。

// TODO: 实现 Flush()
//   使用写锁。清空 data（重新分配一个空 map 或逐个删除）。
