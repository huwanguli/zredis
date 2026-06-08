package core

import "time"

// Expire 设置 key 的过期时间（从现在起 ttl 后过期）。
// 如果 key 不存在，返回 false；否则设置过期并返回 true。
func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.expire(key, ttl)
}

// TTL 返回 key 的剩余生存时间。
// -2: key 不存在
// -1: key 存在但没有设置过期
// 正数: 剩余生存时间
func (s *Store) TTL(key string) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.data[key]; !ok {
		return -2
	}
	exp, hasExp := s.expires[key]
	if !hasExp {
		return -1
	}
	return time.Until(exp)
}

// Persist 移除 key 的过期时间，使之永久存在。
// 如果 key 不存在，返回 false；否则移除过期并返回 true。
func (s *Store) Persist(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.persist(key)
}

// CleanExpired 随机抽取最多 batchSize 个带过期时间的 key，删除其中已过期的。
// 返回实际删除的数量。Go map 的遍历顺序本身是随机的，无需额外随机。
func (s *Store) CleanExpired(batchSize int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	deleted := 0
	count := 0
	for key, exp := range s.expires {
		if count >= batchSize {
			break
		}
		if now.After(exp) {
			delete(s.data, key)
			delete(s.expires, key)
			deleted++
		}
		count++
	}
	return deleted
}
