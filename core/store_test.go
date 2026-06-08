package core

import (
	"sync"
	"testing"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("NewStore returned nil")
	}
	if s.data == nil {
		t.Fatal("Store.data is nil, should be initialized")
	}
}

func TestSetGet(t *testing.T) {
	s := NewStore()
	s.Set("name", "zhangsan")

	val, ok := s.Get("name")
	if !ok {
		t.Error("Get returned false for existing key")
	}
	if val != "zhangsan" {
		t.Errorf("expected 'zhangsan', got '%s'", val)
	}
}

func TestSetOverwrite(t *testing.T) {
	s := NewStore()
	s.Set("key", "v1")
	s.Set("key", "v2")

	val, ok := s.Get("key")
	if !ok {
		t.Error("Get returned false after overwrite")
	}
	if val != "v2" {
		t.Errorf("expected 'v2', got '%s'", val)
	}
}

func TestGetNonExistent(t *testing.T) {
	s := NewStore()
	val, ok := s.Get("nonexistent")
	if ok {
		t.Error("Get returned true for non-existent key")
	}
	if val != "" {
		t.Errorf("expected empty string, got '%s'", val)
	}
}

func TestDel(t *testing.T) {
	s := NewStore()
	s.Set("key", "value")

	existed := s.Del("key")
	if !existed {
		t.Error("Del returned false for existing key")
	}

	_, ok := s.Get("key")
	if ok {
		t.Error("key still exists after Del")
	}
}

func TestDelNonExistent(t *testing.T) {
	s := NewStore()
	existed := s.Del("no-key")
	if existed {
		t.Error("Del returned true for non-existent key")
	}
}

func TestExists(t *testing.T) {
	s := NewStore()

	if s.Exists("a") {
		t.Error("Exists returned true for missing key")
	}

	s.Set("a", "1")
	if !s.Exists("a") {
		t.Error("Exists returned false for existing key")
	}

	s.Del("a")
	if s.Exists("a") {
		t.Error("Exists returned true after Del")
	}
}

func TestKeys(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("c", "3")

	keys := s.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, expected := range []string{"a", "b", "c"} {
		if !keySet[expected] {
			t.Errorf("expected key '%s' not found in Keys()", expected)
		}
	}
}

func TestKeysEmpty(t *testing.T) {
	s := NewStore()
	keys := s.Keys()
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestLen(t *testing.T) {
	s := NewStore()
	if s.Len() != 0 {
		t.Errorf("expected Len 0, got %d", s.Len())
	}

	s.Set("a", "1")
	if s.Len() != 1 {
		t.Errorf("expected Len 1, got %d", s.Len())
	}

	s.Set("b", "2")
	s.Set("c", "3")
	if s.Len() != 3 {
		t.Errorf("expected Len 3, got %d", s.Len())
	}

	s.Del("a")
	if s.Len() != 2 {
		t.Errorf("expected Len 2 after Del, got %d", s.Len())
	}
}

func TestFlush(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("b", "2")
	s.Flush()

	if s.Len() != 0 {
		t.Errorf("expected Len 0 after Flush, got %d", s.Len())
	}
	if len(s.Keys()) != 0 {
		t.Error("Keys() not empty after Flush")
	}
}

// --- 并发测试 ---

func TestConcurrentReads(t *testing.T) {
	s := NewStore()
	s.Set("shared", "value")

	var wg sync.WaitGroup
	numReaders := 100

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, ok := s.Get("shared")
			if !ok || val != "value" {
				t.Errorf("concurrent Get failed: ok=%v val=%s", ok, val)
			}
		}()
	}
	wg.Wait()
}

func TestConcurrentWrites(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	numWriters := 50

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s.Set("key", "value")
			s.Get("key")
			s.Exists("key")
		}(i)
	}
	wg.Wait()
}

func TestConcurrentMixedReadWrite(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	numOps := 100

	// 预填充一些数据
	for i := 0; i < 10; i++ {
		s.Set(string(rune('a'+i)), "x")
	}

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := string(rune('a' + (id % 10)))
			switch id % 3 {
			case 0:
				s.Get(key)
			case 1:
				s.Set(key, "updated")
			case 2:
				s.Exists(key)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentKeys(t *testing.T) {
	s := NewStore()
	for i := 0; i < 20; i++ {
		s.Set(string(rune('a'+i)), "x")
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			keys := s.Keys()
			if len(keys) < 20 {
				t.Errorf("Keys returned %d items, expected at least 20", len(keys))
			}
		}()
	}
	wg.Wait()
}
