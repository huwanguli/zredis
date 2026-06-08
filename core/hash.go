package core

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

// hset 设置 hash 中的 field→value 对。调用者必须持有锁。
// pairs 是成对的 field1, value1, field2, value2, ...
// 返回新增的 field 数量（修改已存在的不计入）。
func (s *Store) hset(key string, pairs ...string) int {
	// TODO: 基于 DataEntry 重写
	//   1. 取 entry，不存在则创建 &DataEntry{Type: DataHash, Hash: make(map[string]string)}
	//   2. 如果已存在但 Type != DataHash，这里简单：创建新 entry 覆盖
	//   3. 遍历 pairs（步长 2），填 entry.Hash，统计新增数量
	return 0
}

// --- 导出方法 ---

// HSet 设置 hash 字段。pairs 是成对的 field, value。
// 返回新增的 field 数量。
func (s *Store) HSet(key string, pairs ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hset(key, pairs...)
}

// HGet 获取 hash 中指定 field 的值。
func (s *Store) HGet(key string, field string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hash, ok := s.getHash(key)
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
	hash, ok := s.getHash(key)
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	hash, ok := s.getHash(key)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(hash))
	for k, v := range hash {
		result[k] = v
	}
	return result
}

// HExists 判断 hash 中指定 field 是否存在。
func (s *Store) HExists(key string, field string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hash, ok := s.getHash(key)
	if !ok {
		return false
	}
	_, exists := hash[field]
	return exists
}

// HLen 返回 hash 中 field 的数量。
func (s *Store) HLen(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hash, ok := s.getHash(key)
	if !ok {
		return 0
	}
	return len(hash)
}

// HKeys 返回 hash 中所有 field 名。
func (s *Store) HKeys(key string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hash, ok := s.getHash(key)
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	hash, ok := s.getHash(key)
	if !ok {
		return nil
	}
	vals := make([]string, 0, len(hash))
	for _, v := range hash {
		vals = append(vals, v)
	}
	return vals
}
