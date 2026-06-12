package core

import (
	"fmt"
	"strconv"
)

// --- 内部方法 ---

// getZSet 获取 zset 类型的跳表，调用者必须持有锁。
func (s *Store) getZSet(key string) (*SkipList, bool) {
	entry, ok := s.get(key)
	if !ok || entry.Type != DataZSet {
		return nil, false
	}
	return entry.ZSet, true
}

// lookupZSet 获取 zset 类型的跳表，内置惰性过期删除和类型检查。
func (s *Store) lookupZSet(key string) (*SkipList, bool) {
	s.expireIfNeeded(key)
	return s.getZSet(key)
}

// --- 导出方法 ---

// ZAdd 添加/更新成员分数。pairs 是 score, member, score, member...
// 返回新增的数量（更新已有成员不计入）。
func (s *Store) ZAdd(key string, pairs ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if !ok {
		entry = &DataEntry{Type: DataZSet, ZSet: NewSkipList()}
		s.data[key] = entry
	} else if entry.Type != DataZSet {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	newCount := 0
	for i := 0; i < len(pairs); i += 2 {
		score, err := strconv.ParseFloat(pairs[i], 64)
		if err != nil {
			return 0, err
		}
		member := pairs[i+1]

		oldScore, existed := entry.ZSet.GetScore(member)
		if existed {
			entry.ZSet.Delete(oldScore, member)
		}
		entry.ZSet.Insert(score, member)
		if !existed {
			newCount++
		}
	}
	return newCount, nil
}

// ZRange 按 score 升序返回 [start, stop] 范围的成员。
func (s *Store) ZRange(key string, start, stop int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.lookupZSet(key)
	if !ok {
		return nil
	}
	return sl.GetRange(int64(start), int64(stop))
}

// ZRank 返回 member 的排名（0-based，升序）。不存在返回 false。
func (s *Store) ZRank(key string, member string) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.lookupZSet(key)
	if !ok {
		return 0, false
	}
	score, found := sl.GetScore(member)
	if !found {
		return 0, false
	}
	return sl.GetRank(score, member), true
}

// ZScore 返回 member 的分数。
func (s *Store) ZScore(key string, member string) (float64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.lookupZSet(key)
	if !ok {
		return 0, false
	}
	return sl.GetScore(member)
}

// ZRem 删除一个或多个成员，返回实际删除的数量。
func (s *Store) ZRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeeded(key)
	entry, ok := s.get(key)
	if !ok || entry.Type != DataZSet {
		return 0, nil
	}

	count := 0
	for _, member := range members {
		score, exists := entry.ZSet.GetScore(member)
		if exists {
			entry.ZSet.Delete(score, member)
			count++
		}
	}
	return count, nil
}

// ZCard 返回有序集合的成员数量。
func (s *Store) ZCard(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.lookupZSet(key)
	if !ok {
		return 0
	}
	return int(sl.Length)
}
