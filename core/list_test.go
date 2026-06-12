package core

import "testing"

// --- LPush / RPush ---

func TestLPush_Basic(t *testing.T) {
	s := NewStore()
	n, err := s.LPush("list", "a", "b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected length 3, got %d", n)
	}

	// LPush a b c → 左端推入，结果应该是 c b a
	val, ok := s.LPop("list")
	if !ok || val != "c" {
		t.Errorf("expected 'c', got '%s'", val)
	}
	val, ok = s.LPop("list")
	if !ok || val != "b" {
		t.Errorf("expected 'b', got '%s'", val)
	}
	val, ok = s.LPop("list")
	if !ok || val != "a" {
		t.Errorf("expected 'a', got '%s'", val)
	}
}

func TestRPush_Basic(t *testing.T) {
	s := NewStore()
	_, err := s.RPush("list", "a", "b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// RPush a b c → 右端推入，左弹出应该是 a b c 顺序
	val, _ := s.LPop("list")
	if val != "a" {
		t.Errorf("expected 'a', got '%s'", val)
	}
}

func TestLPush_NonExistent(t *testing.T) {
	s := NewStore()
	n, err := s.LPush("list", "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

// --- LPop / RPop ---

func TestLPop_Empty(t *testing.T) {
	s := NewStore()
	_, ok := s.LPop("noexist")
	if ok {
		t.Error("LPop should return false for non-existent key")
	}
}

func TestRPop_Basic(t *testing.T) {
	s := NewStore()
	s.LPush("list", "a", "b") // b a

	val, ok := s.RPop("list")
	if !ok || val != "a" {
		t.Errorf("expected 'a', got '%s'", val)
	}
	val, ok = s.RPop("list")
	if !ok || val != "b" {
		t.Errorf("expected 'b', got '%s'", val)
	}
	_, ok = s.RPop("list")
	if ok {
		t.Error("RPop should return false for empty list")
	}
}

// --- LLen ---

func TestLLen_Basic(t *testing.T) {
	s := NewStore()
	if s.LLen("list") != 0 {
		t.Error("LLen should return 0 for non-existent key")
	}

	s.LPush("list", "a", "b")
	if s.LLen("list") != 2 {
		t.Errorf("expected 2, got %d", s.LLen("list"))
	}
}

// --- LRange ---

func TestLRange_Basic(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a", "b", "c", "d", "e")

	vals := s.LRange("list", 1, 3)
	if len(vals) != 3 || vals[0] != "b" || vals[1] != "c" || vals[2] != "d" {
		t.Errorf("expected [b c d], got %v", vals)
	}
}

func TestLRange_NegativeIndex(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a", "b", "c", "d", "e")

	vals := s.LRange("list", -2, -1)
	if len(vals) != 2 || vals[0] != "d" || vals[1] != "e" {
		t.Errorf("expected [d e], got %v", vals)
	}
}

func TestLRange_OutOfBounds(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a", "b")

	vals := s.LRange("list", 0, 10)
	if len(vals) != 2 {
		t.Errorf("expected 2, got %d", len(vals))
	}
}

func TestLRange_NonExistent(t *testing.T) {
	s := NewStore()
	vals := s.LRange("noexist", 0, -1)
	if vals != nil {
		t.Error("LRange should return nil for non-existent key")
	}
}

// --- LIndex ---

func TestLIndex_Basic(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a", "b", "c")

	val, ok := s.LIndex("list", 1)
	if !ok || val != "b" {
		t.Errorf("expected 'b', got '%s'", val)
	}
}

func TestLIndex_Negative(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a", "b", "c")

	val, ok := s.LIndex("list", -1)
	if !ok || val != "c" {
		t.Errorf("expected 'c', got '%s'", val)
	}
}

func TestLIndex_OutOfBounds(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a")

	_, ok := s.LIndex("list", 5)
	if ok {
		t.Error("LIndex should return false for out of bounds")
	}
}

// --- LSet ---

func TestLSet_Basic(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a", "b", "c")

	err := s.LSet("list", 1, "X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, _ := s.LIndex("list", 1)
	if val != "X" {
		t.Errorf("expected 'X', got '%s'", val)
	}
}

func TestLSet_OutOfBounds(t *testing.T) {
	s := NewStore()
	s.RPush("list", "a")

	err := s.LSet("list", 5, "X")
	if err == nil {
		t.Error("LSet should return error for out of bounds")
	}
}

// --- Wrong type ---

func TestList_WrongType(t *testing.T) {
	s := NewStore()
	s.Set("key", "string")

	_, err := s.LPush("key", "x")
	if err == nil {
		t.Error("LPush on string key should return error")
	}

	_, err = s.RPush("key", "x")
	if err == nil {
		t.Error("RPush on string key should return error")
	}
}

// --- 并发 ---

func TestList_Concurrent(t *testing.T) {
	s := NewStore()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			s.LPush("list", "x")
			s.RPush("list", "y")
			s.LLen("list")
			s.LPop("list")
			s.RPop("list")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
