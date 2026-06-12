package core

import (
	"fmt"
	"maps"
)

// --- 内部方法 ---

// getHash 获取 hash 类型的 field→value 映射，调用者必须持有锁。
// key 不存在或类型不匹配返回 nil, false。
func (s *Store) getHash(key string) (map[string]string, bool) {
	entry, ok := s.get(key)
	if !ok || entry.Type != DataHash {
		return nil, false
	}
	return entry.Hash, true
}

// lookupHash 获取 hash 类型的映射，内置惰性过期删除和类型检查。
// 调用者必须持有锁。
func (s *Store) lookupHash(key string) (map[string]string, bool) {
	s.expireIfNeeded(key)
	return s.getHash(key)
}

// hset 设置 hash 中的 field→value 对。调用者必须持有锁。
// pairs 是成对的 field1, value1, field2, value2, ...
// 返回新增的 field 数量（修改已存在的不计入）。
func (s *Store) hset(key string, pairs ...string) (int, error) {
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if !ok {
		entry = &DataEntry{Type: DataHash, Hash: make(map[string]string)}
		s.data[key] = entry
	}
	if ok && entry.Type != DataHash {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	newFields := 0
	for i := 0; i < len(pairs); i += 2 {
		field := pairs[i]
		value := pairs[i+1]
		if _, exists := entry.Hash[field]; !exists {
			newFields++
		}
		entry.Hash[field] = value
	}
	return newFields, nil
}

// --- 导出方法 ---

// HSet 设置 hash 字段。pairs 是成对的 field, value。
// 返回新增的 field 数量。
func (s *Store) HSet(key string, pairs ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hset(key, pairs...)
}

// HGet 获取 hash 中指定 field 的值。
func (s *Store) HGet(key string, field string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, ok := s.lookupHash(key)
	if !ok {
		return "", false
	}
	val, ok := hash[field]
	return val, ok
}

// HDel 删除 hash 中的一个或多个 field，返回实际删除的数量。
func (s *Store) HDel(key string, fields ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, ok := s.lookupHash(key)
	if !ok {
		return 0
	}
	deleted := 0
	for _, f := range fields {
		if _, exists := hash[f]; exists {
			delete(hash, f)
			deleted++
		}
	}
	return deleted
}

// HGetAll 返回 hash 全部 field→value 对。key 不存在或类型不匹配返回 nil。
func (s *Store) HGetAll(key string) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, ok := s.lookupHash(key)
	if !ok {
		return nil
	}
	return maps.Clone(hash)
}

// HExists 判断 hash 中指定 field 是否存在。
func (s *Store) HExists(key string, field string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, ok := s.lookupHash(key)
	if !ok {
		return false
	}
	_, exists := hash[field]
	return exists
}

// HLen 返回 hash 中 field 的数量。
func (s *Store) HLen(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, ok := s.lookupHash(key)
	if !ok {
		return 0
	}
	return len(hash)
}

// HKeys 返回 hash 中所有 field 名。
func (s *Store) HKeys(key string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, ok := s.lookupHash(key)
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(hash))
	for k := range hash {
		keys = append(keys, k)
	}
	return keys
}

// HVals 返回 hash 中所有 value。
func (s *Store) HVals(key string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	hash, ok := s.lookupHash(key)
	if !ok {
		return nil
	}
	vals := make([]string, 0, len(hash))
	for _, v := range hash {
		vals = append(vals, v)
	}
	return vals
}
