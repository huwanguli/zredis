package core

import (
	"sync"
	"time"
)

// Store 是一个线程安全的内存键值存储引擎。
type Store struct {
	mu      sync.RWMutex
	data    map[string]string
	expires map[string]time.Time
}

// NewStore 创建并初始化一个新的 Store。
func NewStore() *Store {
	return &Store{
		data:    make(map[string]string),
		expires: make(map[string]time.Time),
	}
}

// --- 内部方法（小写，不加锁，由外部方法调用） ---

// get 获取值（不检查过期），调用者必须持有锁。
func (s *Store) get(key string) (string, bool) {
	val, ok := s.data[key]
	return val, ok
}

// set 设置值，调用者必须持有锁。
func (s *Store) set(key string, value string) {
	s.data[key] = value
}

// del 删除 key 及其过期记录，返回是否原本存在。调用者必须持有锁。
func (s *Store) del(key string) bool {
	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		delete(s.expires, key)
		return true
	}
	return false
}

// expire 设置过期时间，调用者必须持有锁。key 不存在返回 false。
func (s *Store) expire(key string, ttl time.Duration) bool {
	if _, ok := s.data[key]; ok {
		s.expires[key] = time.Now().Add(ttl)
		return true
	}
	return false
}

// persist 移除过期时间，调用者必须持有锁。
func (s *Store) persist(key string) bool {
	if _, ok := s.data[key]; ok {
		delete(s.expires, key)
		return true
	}
	return false
}

// --- 导出方法（大写，加锁后调内部方法） ---

// Get 获取 key 对应的值。使用写锁以支持惰性过期删除。
// 如果 key 不存在或已过期，返回 "", false。
func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.get(key)
	if !ok {
		return "", false
	}
	if exp, hasExp := s.expires[key]; hasExp && time.Now().After(exp) {
		s.del(key)
		return "", false
	}
	return val, ok
}

// Set 设置键值对，覆盖已有值。
func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set(key, value)
}

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
	s.data = make(map[string]string)
	s.expires = make(map[string]time.Time)
}
