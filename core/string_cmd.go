package core

import (
	"maps"
	"strconv"
	"time"
)

// incrBy 将 key 的值加 delta 并返回新值。调用者必须持有锁。
// key 不存在时视为 0；值非数字时报错。
func (s *Store) incrBy(key string, delta int64) (int64, error) {
	val, ok := s.get(key)
	if !ok {
		val = "0"
	}
	current, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	newValue := current + delta
	s.data[key] = strconv.FormatInt(newValue, 10)
	return newValue, nil
}

// --- INCR / INCRBY / DECR / DECRBY ---

// Incr 将 key 的值加 1 并返回新值。
func (s *Store) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.incrBy(key, 1)
}

// IncrBy 将 key 的值加 delta 并返回新值。
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
// 如果 key 不存在，相当于 Set。
func (s *Store) Append(key string, value string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] += value
	return len(s.data[key])
}

// --- STRLEN ---

// StrLen 返回 key 对应字符串的字节长度。key 不存在返回 0。
func (s *Store) StrLen(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data[key])
}

// --- MGET / MSET ---

// MGet 批量获取多个 key 的值，返回与 keys 一一对应的 BulkVal 切片。
// key 存在时为 NewBulkVal，不存在时为 NewNullBulkVal。
func (s *Store) MGet(keys ...string) []*BulkVal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*BulkVal, len(keys))
	for i, key := range keys {
		if val, ok := s.data[key]; ok {
			results[i] = NewBulkVal([]byte(val))
		} else {
			results[i] = NewNullBulkVal()
		}
	}
	return results
}

// MSet 批量设置键值对，所有 key 在一次锁内原子写入。
func (s *Store) MSet(kvs map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	maps.Copy(s.data, kvs)
}

// --- GETSET ---

// GetSet 设置新值并返回旧值。如果 key 不存在，返回 "", false。
func (s *Store) GetSet(key string, value string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldValue, ok := s.data[key]
	s.set(key, value)
	if !ok {
		return "", false
	}
	return oldValue, true
}

// --- SETEX ---

// SetEX 设置值并同时设置过期时间，保证 Set + Expire 原子性。
func (s *Store) SetEX(key string, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.set(key, value)
	s.expire(key, ttl)
}

// --- SETNX ---

// SetNX 仅当 key 不存在时设置值。key 已存在则不做任何操作，返回 false。
func (s *Store) SetNX(key string, value string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[key]; exists {
		return false
	}
	s.set(key, value)
	return true
}
