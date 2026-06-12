package core

import (
	"fmt"
	"strconv"
	"time"
)

// --- String 专属内部方法 ---

// getString 获取 string 类型的 DataEntry。调用者必须持有锁。
// key 不存在或类型不匹配返回 nil, false。
func (s *Store) getString(key string) (*DataEntry, bool) {
	entry, ok := s.get(key)
	if !ok || entry.Type != DataString {
		return nil, false
	}
	return entry, true
}

// setString 设置 key 为 string 类型值，调用者必须持有锁。
func (s *Store) setString(key string, value string) {
	s.data[key] = &DataEntry{Type: DataString, String: value}
}

// lookupString 获取 string 类型的 DataEntry，内置惰性过期删除和类型检查。
// 调用者必须持有锁。
func (s *Store) lookupString(key string) (*DataEntry, bool) {
	s.expireIfNeeded(key)
	return s.getString(key)
}

// incrBy 将 key 的值加 delta。调用者必须持有锁。
// key 不存在时视作 0；类型不匹配报错。
func (s *Store) incrBy(key string, delta int64) (int64, error) {
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if !ok {
		entry = &DataEntry{Type: DataString, String: "0"}
		s.data[key] = entry
	}
	if entry.Type != DataString {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	current, err := strconv.ParseInt(entry.String, 10, 64)
	if err != nil {
		return 0, err
	}
	newValue := current + delta
	entry.String = strconv.FormatInt(newValue, 10)
	return newValue, nil
}

// --- GET / SET ---

// Get 获取 key 对应的 string 值。使用写锁以支持惰性过期删除。
func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.lookupString(key)
	if !ok {
		return "", false
	}
	return entry.String, true
}

// Set 设置 key 为 string 类型值，覆盖已有值。
func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setString(key, value)
}

// --- INCR / INCRBY / DECR / DECRBY ---

// Incr 将 key 的值加 1。
func (s *Store) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.incrBy(key, 1)
}

// IncrBy 将 key 的值加 delta。
func (s *Store) IncrBy(key string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.incrBy(key, delta)
}

// Decr 将 key 的值减 1。
func (s *Store) Decr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.incrBy(key, -1)
}

// DecrBy 将 key 的值减 delta。
func (s *Store) DecrBy(key string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.incrBy(key, -delta)
}

// --- APPEND ---

// Append 将 value 追加到 key 已有值的末尾，返回追加后的总长度。
// key 不存在相当于新建；类型不匹配返回错误。
func (s *Store) Append(key string, value string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.lookupString(key)
	if !ok {
		if _, exists := s.get(key); exists {
			return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		s.setString(key, value)
		return len(value), nil
	}
	entry.String += value
	return len(entry.String), nil
}

// --- STRLEN ---

// StrLen 返回 string 类型 key 的字节长度。key 不存在、已过期或类型不匹配返回 0。
func (s *Store) StrLen(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.lookupString(key)
	if !ok {
		return 0
	}
	return len(entry.String)
}

// --- MGET / MSET ---

// MGet 批量获取多个 key 的 string 值，返回与 keys 一一对应的结果。
// key 不存在、已过期或类型不匹配返回空字符串。
func (s *Store) MGet(keys ...string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	results := make([]string, len(keys))
	for i, key := range keys {
		entry, ok := s.lookupString(key)
		if ok {
			results[i] = entry.String
		}
	}
	return results
}

// MSet 批量设置 string 键值对，原子写入。
func (s *Store) MSet(kvs map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range kvs {
		s.setString(key, value)
	}
}

// --- GETSET ---

// GetSet 设置新值并返回旧值。key 不存在或类型不匹配返回 "", false。
func (s *Store) GetSet(key string, value string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.lookupString(key)
	oldVal := ""
	if ok {
		oldVal = entry.String
	}
	s.setString(key, value)
	return oldVal, ok
}

// --- SETEX ---

// SetEX 设置值并同时设置过期时间，保证原子性。
func (s *Store) SetEX(key string, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setString(key, value)
	s.expire(key, ttl)
}

// --- SETNX ---

// SetNX 仅当 key 不存在时设置值。key 已存在返回 false。
func (s *Store) SetNX(key string, value string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeeded(key)
	if _, exists := s.get(key); exists {
		return false
	}
	s.setString(key, value)
	return true
}
