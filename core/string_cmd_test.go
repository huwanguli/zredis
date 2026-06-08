package core

import (
	"strconv"
	"testing"
	"time"
)

// --- INCR ---

func TestIncr_NonExistent(t *testing.T) {
	s := NewStore()
	n, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestIncr_Existing(t *testing.T) {
	s := NewStore()
	s.Set("counter", "10")
	n, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 11 {
		t.Errorf("expected 11, got %d", n)
	}
}

func TestIncr_NotANumber(t *testing.T) {
	s := NewStore()
	s.Set("counter", "hello")
	_, err := s.Incr("counter")
	if err == nil {
		t.Error("expected error for non-numeric value")
	}
}

func TestIncr_Negative(t *testing.T) {
	s := NewStore()
	s.Set("counter", "-5")
	n, err := s.Incr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != -4 {
		t.Errorf("expected -4, got %d", n)
	}
}

// --- INCRBY ---

func TestIncrBy_Existing(t *testing.T) {
	s := NewStore()
	s.Set("key", "100")
	n, err := s.IncrBy("key", 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 125 {
		t.Errorf("expected 125, got %d", n)
	}
}

func TestIncrBy_NegativeDelta(t *testing.T) {
	s := NewStore()
	s.Set("key", "50")
	n, err := s.IncrBy("key", -30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 20 {
		t.Errorf("expected 20, got %d", n)
	}
}

// --- DECR ---

func TestDecr_Basic(t *testing.T) {
	s := NewStore()
	s.Set("counter", "5")
	n, err := s.Decr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
}

func TestDecr_NonExistent(t *testing.T) {
	s := NewStore()
	n, err := s.Decr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != -1 {
		t.Errorf("expected -1, got %d", n)
	}
}

// --- DECRBY ---

func TestDecrBy_Basic(t *testing.T) {
	s := NewStore()
	s.Set("key", "100")
	n, err := s.DecrBy("key", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 70 {
		t.Errorf("expected 70, got %d", n)
	}
}

// --- APPEND ---

func TestAppend_Existing(t *testing.T) {
	s := NewStore()
	s.Set("key", "hello")
	n := s.Append("key", " world")
	if n != 11 {
		t.Errorf("expected length 11, got %d", n)
	}
	val, ok := s.Get("key")
	if !ok || val != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", val)
	}
}

func TestAppend_NonExistent(t *testing.T) {
	s := NewStore()
	n := s.Append("key", "hello")
	if n != 5 {
		t.Errorf("expected length 5, got %d", n)
	}
	val, _ := s.Get("key")
	if val != "hello" {
		t.Errorf("expected 'hello', got '%s'", val)
	}
}

func TestAppend_Empty(t *testing.T) {
	s := NewStore()
	s.Set("key", "")
	n := s.Append("key", "data")
	if n != 4 {
		t.Errorf("expected length 4, got %d", n)
	}
}

// --- STRLEN ---

func TestStrLen_Existing(t *testing.T) {
	s := NewStore()
	s.Set("key", "hello")
	if n := s.StrLen("key"); n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
}

func TestStrLen_NonExistent(t *testing.T) {
	s := NewStore()
	if n := s.StrLen("nope"); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestStrLen_Empty(t *testing.T) {
	s := NewStore()
	s.Set("key", "")
	if n := s.StrLen("key"); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestStrLen_Unicode(t *testing.T) {
	s := NewStore()
	s.Set("key", "你好")
	// "你好" is 6 bytes in UTF-8 (3 bytes each)
	if n := s.StrLen("key"); n != 6 {
		t.Errorf("expected 6, got %d", n)
	}
}

// --- MGET ---

func TestMGet_AllExist(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("c", "3")

	vals := s.MGet("a", "b", "c")
	if len(vals) != 3 {
		t.Fatalf("expected 3 results, got %d", len(vals))
	}
	if vals[0] != "1" {
		t.Errorf("MGet[0] expected '1', got '%s'", vals[0])
	}
	if vals[1] != "2" {
		t.Errorf("MGet[1] expected '2', got '%s'", vals[1])
	}
	if vals[2] != "3" {
		t.Errorf("MGet[2] expected '3', got '%s'", vals[2])
	}
}

func TestMGet_Mixed(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	// "b" doesn't exist

	vals := s.MGet("a", "b")
	if len(vals) != 2 {
		t.Fatalf("expected 2 results, got %d", len(vals))
	}
	if vals[0] != "1" {
		t.Errorf("MGet[0] expected '1', got '%s'", vals[0])
	}
	if vals[1] != "" {
		t.Errorf("MGet[1] expected '', got '%s'", vals[1])
	}
}

func TestMGet_AllMissing(t *testing.T) {
	s := NewStore()
	vals := s.MGet("x", "y")
	for i, v := range vals {
		if v != "" {
			t.Errorf("MGet[%d] expected '', got '%s'", i, v)
		}
	}
}

// --- MSET ---

func TestMSet_Basic(t *testing.T) {
	s := NewStore()
	s.MSet(map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	})

	if s.Len() != 3 {
		t.Errorf("expected 3 keys, got %d", s.Len())
	}
	val, _ := s.Get("b")
	if val != "2" {
		t.Errorf("expected '2', got '%s'", val)
	}
}

func TestMSet_Overwrite(t *testing.T) {
	s := NewStore()
	s.Set("a", "old")
	s.MSet(map[string]string{"a": "new", "b": "x"})

	val, _ := s.Get("a")
	if val != "new" {
		t.Errorf("expected 'new', got '%s'", val)
	}
}

// --- GETSET ---

func TestGetSet_Existing(t *testing.T) {
	s := NewStore()
	s.Set("key", "old")

	oldVal, ok := s.GetSet("key", "new")
	if !ok {
		t.Error("GetSet should return true for existing key")
	}
	if oldVal != "old" {
		t.Errorf("expected old value 'old', got '%s'", oldVal)
	}

	newVal, _ := s.Get("key")
	if newVal != "new" {
		t.Errorf("expected new value 'new', got '%s'", newVal)
	}
}

func TestGetSet_NonExistent(t *testing.T) {
	s := NewStore()
	oldVal, ok := s.GetSet("key", "value")
	if ok {
		t.Error("GetSet should return false for non-existent key")
	}
	if oldVal != "" {
		t.Errorf("expected empty, got '%s'", oldVal)
	}

	val, exists := s.Get("key")
	if !exists || val != "value" {
		t.Error("value should be set")
	}
}

// --- SETEX ---

func TestSetEx_Basic(t *testing.T) {
	s := NewStore()
	s.SetEX("key", "value", time.Hour)

	val, ok := s.Get("key")
	if !ok || val != "value" {
		t.Error("SetEX should set the value")
	}

	ttl := s.TTL("key")
	if ttl <= 0 {
		t.Error("SetEX should set TTL")
	}
}

func TestSetEx_Expires(t *testing.T) {
	s := NewStore()
	s.SetEX("key", "value", 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	_, ok := s.Get("key")
	if ok {
		t.Error("key should have expired")
	}
}

// --- SETNX ---

func TestSetNX_NonExistent(t *testing.T) {
	s := NewStore()
	ok := s.SetNX("key", "value")
	if !ok {
		t.Error("SetNX should return true for new key")
	}
	val, _ := s.Get("key")
	if val != "value" {
		t.Errorf("expected 'value', got '%s'", val)
	}
}

func TestSetNX_AlreadyExists(t *testing.T) {
	s := NewStore()
	s.Set("key", "original")
	ok := s.SetNX("key", "new")
	if ok {
		t.Error("SetNX should return false for existing key")
	}
	val, _ := s.Get("key")
	if val != "original" {
		t.Error("existing value should not be overwritten")
	}
}

// --- 并发 ---

func TestIncr_Concurrent(t *testing.T) {
	s := NewStore()
	done := make(chan bool)
	n := 100

	for i := 0; i < n; i++ {
		go func() {
			_, err := s.Incr("counter")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < n; i++ {
		<-done
	}

	val, _ := s.Get("counter")
	if val != strconv.Itoa(n) {
		t.Errorf("expected %d, got %s", n, val)
	}
}
