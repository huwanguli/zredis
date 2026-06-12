package core

import "testing"

// --- HSet / HGet ---

func TestHSet_HGet_Basic(t *testing.T) {
	s := NewStore()
	n, _  := s.HSet("user:1", "name", "zhangsan", "age", "25")
	if n != 2 {
		t.Errorf("expected 2 new fields, got %d", n)
	}

	val, ok := s.HGet("user:1", "name")
	if !ok || val != "zhangsan" {
		t.Errorf("expected zhangsan, got %s (ok=%v)", val, ok)
	}

	val, ok = s.HGet("user:1", "age")
	if !ok || val != "25" {
		t.Errorf("expected 25, got %s (ok=%v)", val, ok)
	}
}

func TestHSet_UpdateExisting(t *testing.T) {
	s := NewStore()
	s.HSet("key", "a", "1")
	n,_ := s.HSet("key", "a", "new", "b", "2")
	if n != 1 {
		t.Errorf("expected 1 new field (b), got %d", n)
	}

	val, _ := s.HGet("key", "a")
	if val != "new" {
		t.Errorf("expected 'new', got '%s'", val)
	}
}

func TestHGet_NonExistentKey(t *testing.T) {
	s := NewStore()
	_, ok := s.HGet("no-key", "field")
	if ok {
		t.Error("HGet should return false for non-existent key")
	}
}

func TestHGet_NonExistentField(t *testing.T) {
	s := NewStore()
	s.HSet("key", "a", "1")
	_, ok := s.HGet("key", "no-field")
	if ok {
		t.Error("HGet should return false for non-existent field")
	}
}

// --- HDel ---

func TestHDel_Basic(t *testing.T) {
	s := NewStore()
	s.HSet("key", "a", "1", "b", "2", "c", "3")

	n := s.HDel("key", "a", "c")
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}

	_, ok := s.HGet("key", "a")
	if ok {
		t.Error("field 'a' should be deleted")
	}
	_, ok = s.HGet("key", "b")
	if !ok {
		t.Error("field 'b' should still exist")
	}
}

func TestHDel_NonExistentKey(t *testing.T) {
	s := NewStore()
	n := s.HDel("no-key", "field")
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

// --- HGetAll ---

func TestHGetAll_Basic(t *testing.T) {
	s := NewStore()
	s.HSet("key", "a", "1", "b", "2")

	all := s.HGetAll("key")
	if len(all) != 2 {
		t.Errorf("expected 2 fields, got %d", len(all))
	}
	if all["a"] != "1" || all["b"] != "2" {
		t.Error("HGetAll returned wrong values")
	}
}

func TestHGetAll_NonExistent(t *testing.T) {
	s := NewStore()
	all := s.HGetAll("no-key")
	if all != nil {
		t.Error("HGetAll should return nil for non-existent key")
	}
}

func TestHGetAll_Empty(t *testing.T) {
	s := NewStore()
	s.HSet("key", "a", "1")
	s.HDel("key", "a")

	all := s.HGetAll("key")
	if len(all) != 0 {
		t.Errorf("expected empty map, got %d fields", len(all))
	}
}

// --- HExists ---

func TestHExists_Basic(t *testing.T) {
	s := NewStore()
	s.HSet("key", "a", "1")

	if !s.HExists("key", "a") {
		t.Error("HExists should return true")
	}
	if s.HExists("key", "b") {
		t.Error("HExists should return false for missing field")
	}
	if s.HExists("no-key", "a") {
		t.Error("HExists should return false for missing key")
	}
}

// --- HLen ---

func TestHLen_Basic(t *testing.T) {
	s := NewStore()
	if s.HLen("key") != 0 {
		t.Error("HLen should return 0 for non-existent key")
	}

	s.HSet("key", "a", "1", "b", "2")
	if s.HLen("key") != 2 {
		t.Errorf("expected 2, got %d", s.HLen("key"))
	}
}

// --- HKeys ---

func TestHKeys_Basic(t *testing.T) {
	s := NewStore()
	keys := s.HKeys("no-key")
	if keys != nil {
		t.Error("HKeys should return nil for non-existent key")
	}

	s.HSet("key", "a", "1", "b", "2")
	keys = s.HKeys("key")
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

// --- HVals ---

func TestHVals_Basic(t *testing.T) {
	s := NewStore()
	vals := s.HVals("no-key")
	if vals != nil {
		t.Error("HVals should return nil for non-existent key")
	}

	s.HSet("key", "a", "1", "b", "2")
	vals = s.HVals("key")
	if len(vals) != 2 {
		t.Errorf("expected 2 vals, got %d", len(vals))
	}
}

// --- 并发 ---

func TestHash_Concurrent(t *testing.T) {
	s := NewStore()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			key := string(rune('a' + id))
			s.HSet(key, "f", "v")
			s.HGet(key, "f")
			s.HExists(key, "f")
			s.HLen(key)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
