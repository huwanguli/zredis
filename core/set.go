package core

import "fmt"

// --- 内部方法 ---

// getSet 获取 set 类型的集合，调用者必须持有锁。
func (s *Store) getSet(key string) (map[string]struct{}, bool) {
	entry, ok := s.get(key)
	if !ok || entry.Type != DataSet {
		return nil, false
	}
	return entry.Set, true
}

// lookupSet 获取 set 类型的集合，内置惰性过期删除和类型检查。
func (s *Store) lookupSet(key string) (map[string]struct{}, bool) {
	s.expireIfNeeded(key)
	return s.getSet(key)
}

// sadd 添加一个或多个成员。调用者必须持有锁，且已确保类型正确。
func (s *Store) sadd(key string, members ...string) int {
	entry, ok := s.get(key)
	if !ok {
		entry = &DataEntry{Type: DataSet, Set: make(map[string]struct{})}
		s.data[key] = entry
	}
	count := 0
	for _, member := range members {
		if _, exists := entry.Set[member]; !exists {
			entry.Set[member] = struct{}{}
			count++
		}
	}
	return count
}

// srem 删除一个或多个成员。调用者必须持有锁，且已确保类型正确。
func (s *Store) srem(key string, members ...string) int {
	entry, ok := s.get(key)
	if !ok {
		return 0
	}
	count := 0
	for _, member := range members {
		if _, exists := entry.Set[member]; exists {
			delete(entry.Set, member)
			count++
		}
	}
	return count
}

// --- 导出方法 ---

// SAdd 添加一个或多个成员到集合，返回新增的数量。
func (s *Store) SAdd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if ok && entry.Type != DataSet {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return s.sadd(key, members...), nil
}

// SRem 从集合中删除一个或多个成员，返回实际删除的数量。
func (s *Store) SRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if ok && entry.Type != DataSet {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return s.srem(key, members...), nil
}

// SMembers 返回集合中所有成员。
func (s *Store) SMembers(key string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	set, ok := s.lookupSet(key)
	if !ok {
		return nil
	}
	members := make([]string, 0, len(set))
	for member := range set {
		members = append(members, member)
	}
	return members
}

// SIsMember 判断 member 是否是集合的成员。
func (s *Store) SIsMember(key string, member string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	set, ok := s.lookupSet(key)
	if !ok {
		return false
	}
	_, exists := set[member]
	return exists
}

// SCard 返回集合的成员数量。
func (s *Store) SCard(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	set, ok := s.lookupSet(key)
	if !ok {
		return 0
	}
	return len(set)
}

// SInter 返回两个集合的交集。
func (s *Store) SInter(key1, key2 string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	set1, ok1 := s.lookupSet(key1)
	set2, ok2 := s.lookupSet(key2)
	if !ok1 || !ok2 {
		return nil
	}
	if len(set1) > len(set2) {
		set1, set2 = set2, set1
	}
	result := make([]string, 0)
	for member := range set1 {
		if _, exists := set2[member]; exists {
			result = append(result, member)
		}
	}
	return result
}

// SUnion 返回两个集合的并集。
func (s *Store) SUnion(key1, key2 string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	set1, ok1 := s.lookupSet(key1)
	set2, ok2 := s.lookupSet(key2)
	if !ok1 && !ok2 {
		return nil
	}
	union := make(map[string]struct{})
	if ok1 {
		for member := range set1 {
			union[member] = struct{}{}
		}
	}
	if ok2 {
		for member := range set2 {
			union[member] = struct{}{}
		}
	}
	result := make([]string, 0, len(union))
	for member := range union {
		result = append(result, member)
	}
	return result
}

// SDiff 返回 key1 - key2 的差集。
func (s *Store) SDiff(key1, key2 string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	set1, ok1 := s.lookupSet(key1)
	set2, ok2 := s.lookupSet(key2)
	if !ok1 || !ok2 {
		return nil
	}
	result := make([]string, 0)
	for member := range set1 {
		if _, exists := set2[member]; !exists {
			result = append(result, member)
		}
	}
	return result
}
