package core

import (
	"testing"
	"time"
)

// --- 惰性过期 ---

func TestGet_LazyExpiration(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")
	s.Expire("key", 10*time.Millisecond)

	// 等 key 过期
	time.Sleep(20 * time.Millisecond)

	val, ok := s.Get("key")
	if ok {
		t.Error("Get should return false for expired key")
	}
	if val != "" {
		t.Errorf("expected empty string, got '%s'", val)
	}
}

func TestGet_ExpiredKeyRemoved(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")
	s.Expire("key", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	s.Get("key") // 触发惰性删除

	if s.Exists("key") {
		t.Error("expired key should be removed from data after Get")
	}
	if s.Len() != 0 {
		t.Errorf("Len should be 0 after expired key removed, got %d", s.Len())
	}
}

func TestGet_NoExpiryValue(t *testing.T) {
	s := NewStore()
	s.Set("permanent", "forever")

	val, ok := s.Get("permanent")
	if !ok {
		t.Error("Get returned false for permanent key")
	}
	if val != "forever" {
		t.Errorf("expected 'forever', got '%s'", val)
	}
}

// --- Expire ---

func TestExpire_ExistingKey(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")

	ok := s.Expire("key", time.Hour)
	if !ok {
		t.Error("Expire should return true for existing key")
	}
}

func TestExpire_NonExistentKey(t *testing.T) {
	s := NewStore()

	ok := s.Expire("no-key", time.Hour)
	if ok {
		t.Error("Expire should return false for non-existent key")
	}
}

// --- TTL ---

func TestTTL_NoExpiry(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")

	ttl := s.TTL("key")
	if ttl != -1 {
		t.Errorf("TTL for permanent key should be -1, got %v", ttl)
	}
}

func TestTTL_NonExistent(t *testing.T) {
	s := NewStore()

	ttl := s.TTL("no-key")
	if ttl != -2 {
		t.Errorf("TTL for non-existent key should be -2, got %v", ttl)
	}
}

func TestTTL_WithExpiry(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")
	s.Expire("key", 10*time.Second)

	ttl := s.TTL("key")
	if ttl <= 0 {
		t.Errorf("TTL should be positive, got %v", ttl)
	}
	if ttl > 10*time.Second {
		t.Errorf("TTL should be <= 10s, got %v", ttl)
	}
}

// --- Persist ---

func TestPersist_ExistingKey(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")
	s.Expire("key", time.Hour)

	ok := s.Persist("key")
	if !ok {
		t.Error("Persist should return true for existing key")
	}

	ttl := s.TTL("key")
	if ttl != -1 {
		t.Errorf("after Persist, TTL should be -1, got %v", ttl)
	}
}

func TestPersist_NonExistentKey(t *testing.T) {
	s := NewStore()

	ok := s.Persist("no-key")
	if ok {
		t.Error("Persist should return false for non-existent key")
	}
}

// --- CleanExpired ---

func TestCleanExpired_Basic(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("c", "3")

	s.Expire("a", time.Millisecond)
	s.Expire("b", time.Millisecond)
	// "c" 没有过期时间

	time.Sleep(10 * time.Millisecond)

	deleted := s.CleanExpired(10)
	if deleted != 2 {
		t.Errorf("CleanExpired should delete 2 keys, got %d", deleted)
	}

	if s.Exists("a") {
		t.Error("'a' should be deleted")
	}
	if s.Exists("b") {
		t.Error("'b' should be deleted")
	}
	if !s.Exists("c") {
		t.Error("'c' should still exist")
	}
}

func TestCleanExpired_LimitedBatch(t *testing.T) {
	s := NewStore()
	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		s.Set(key, "x")
		s.Expire(key, time.Millisecond)
	}

	time.Sleep(10 * time.Millisecond)

	deleted := s.CleanExpired(3)
	if deleted > 3 {
		t.Errorf("CleanExpired should delete at most 3 keys, got %d", deleted)
	}
}

// --- Flush ---

func TestFlush_AlsoClearsExpires(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")
	s.Expire("key", time.Hour)
	s.Flush()

	if s.Len() != 0 {
		t.Error("Len should be 0 after Flush")
	}
	ttl := s.TTL("key")
	if ttl != -2 {
		t.Errorf("TTL after Flush should be -2, got %v", ttl)
	}
}

// --- Del ---

func TestDel_AlsoClearsExpiry(t *testing.T) {
	s := NewStore()
	s.Set("key", "val")
	s.Expire("key", time.Hour)
	s.Del("key")

	ttl := s.TTL("key")
	if ttl != -2 {
		t.Errorf("TTL after Del should be -2, got %v", ttl)
	}
}

// --- 并发安全 ---

func TestTTL_Concurrent(t *testing.T) {
	s := NewStore()
	for i := 0; i < 50; i++ {
		key := string(rune('a' + (i % 26)))
		s.Set(key, "x")
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				key := string(rune('a' + (j % 26)))
				s.Expire(key, time.Hour)
				s.TTL(key)
				s.Persist(key)
				s.CleanExpired(5)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
