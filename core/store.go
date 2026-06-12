package core

import (
	"container/list"
	"sync"
	"time"
)

// DataType 标识一个 key 存储的数据类型。
type DataType byte

const (
	DataString DataType = iota
	DataHash
	DataList
)

// DataEntry 是 store 中实际存储的值，包含类型标记和对应数据。
type DataEntry struct {
	Type   DataType
	String string
	Hash   map[string]string
	List   *list.List
}

// Store 是一个线程安全的内存键值存储引擎。
type Store struct {
	mu      sync.RWMutex
	data    map[string]*DataEntry
	expires map[string]time.Time
}

// NewStore 创建并初始化一个新的 Store。
func NewStore() *Store {
	return &Store{
		data:    make(map[string]*DataEntry),
		expires: make(map[string]time.Time),
	}
}

// --- 通用内部方法（类型无关） ---

// get 获取 key 对应的 DataEntry，不存在返回 nil, false。
func (s *Store) get(key string) (*DataEntry, bool) {
	entry, ok := s.data[key]
	if !ok {
		return nil, false
	}
	return entry, true
}

// del 删除 key 及其过期记录，返回是否原本存在。
func (s *Store) del(key string) bool {
	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		delete(s.expires, key)
		return true
	}
	return false
}

// expire 设置过期时间，key 不存在返回 false。
func (s *Store) expire(key string, ttl time.Duration) bool {
	if _, ok := s.data[key]; ok {
		s.expires[key] = time.Now().Add(ttl)
		return true
	}
	return false
}

// persist 移除过期时间。
func (s *Store) persist(key string) bool {
	if _, ok := s.data[key]; ok {
		delete(s.expires, key)
		return true
	}
	return false
}

// expireIfNeeded 惰性删除：如果 key 已过期则删除，返回 true 表示 key 已不存在。
// 调用者必须持有锁。
func (s *Store) expireIfNeeded(key string) bool {
	exp, hasExp := s.expires[key]
	if !hasExp {
		return false
	}
	if time.Now().After(exp) {
		s.del(key)
		return true
	}
	return false
}

// --- 通用导出方法 ---

// Del 删除 key 及其过期记录，返回 key 是否原本存在。
func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.del(key)
}

// Exists 判断 key 是否存在（不检查过期）。
func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

// Keys 返回 store 中所有 key 的列表。
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Len 返回 store 中 key 的数量。
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Flush 清空所有数据和过期记录。
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]*DataEntry)
	s.expires = make(map[string]time.Time)
}
