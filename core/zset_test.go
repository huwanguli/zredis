package core

import "testing"

// --- ZAdd ---

func TestZAdd_Basic(t *testing.T) {
	s := NewStore()
	n, err := s.ZAdd("z", "10", "a", "20", "b", "30", "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}

func TestZAdd_Update(t *testing.T) {
	s := NewStore()
	s.ZAdd("z", "10", "a", "20", "b")
	n, _ := s.ZAdd("z", "15", "a", "25", "c")
	if n != 1 {
		t.Errorf("expected 1 new member, got %d", n)
	}
	score, _ := s.ZScore("z", "a")
	if score != 15 {
		t.Errorf("expected score 15, got %v", score)
	}
}

// --- ZRange ---

func TestZRange_Basic(t *testing.T) {
	s := NewStore()
	s.ZAdd("z", "30", "c", "10", "a", "20", "b")

	members := s.ZRange("z", 0, -1)
	if len(members) != 3 || members[0] != "a" || members[1] != "b" || members[2] != "c" {
		t.Errorf("expected [a b c], got %v", members)
	}
}

func TestZRange_NegativeIndex(t *testing.T) {
	s := NewStore()
	s.ZAdd("z", "10", "a", "20", "b", "30", "c")

	members := s.ZRange("z", -2, -1)
	if len(members) != 2 || members[0] != "b" || members[1] != "c" {
		t.Errorf("expected [b c], got %v", members)
	}
}

func TestZRange_SameScore(t *testing.T) {
	s := NewStore()
	s.ZAdd("z", "10", "b", "10", "a")

	members := s.ZRange("z", 0, -1)
	if members[0] != "a" || members[1] != "b" {
		t.Errorf("same score should sort alphabetically, got %v", members)
	}
}

// --- ZRank ---

func TestZRank_Basic(t *testing.T) {
	s := NewStore()
	s.ZAdd("z", "10", "a", "20", "b", "30", "c")

	rank, ok := s.ZRank("z", "b")
	if !ok || rank != 1 {
		t.Errorf("expected rank 1, got %d (ok=%v)", rank, ok)
	}

	_, ok = s.ZRank("z", "d")
	if ok {
		t.Error("ZRank should return false for non-existent member")
	}
}

// --- ZScore ---

func TestZScore_Basic(t *testing.T) {
	s := NewStore()
	s.ZAdd("z", "10", "a", "20", "b")

	score, ok := s.ZScore("z", "a")
	if !ok || score != 10 {
		t.Errorf("expected 10, got %v", score)
	}
}

// --- ZRem ---

func TestZRem_Basic(t *testing.T) {
	s := NewStore()
	s.ZAdd("z", "10", "a", "20", "b", "30", "c")
	n, _ := s.ZRem("z", "a", "c")
	if n != 2 {
		t.Errorf("expected 2, got %d", n)
	}
	if s.ZCard("z") != 1 {
		t.Errorf("expected 1 remaining, got %d", s.ZCard("z"))
	}
}

// --- ZCard ---

func TestZCard_Basic(t *testing.T) {
	s := NewStore()
	if s.ZCard("z") != 0 {
		t.Error("ZCard should return 0 for non-existent key")
	}
	s.ZAdd("z", "10", "a", "20", "b")
	if s.ZCard("z") != 2 {
		t.Errorf("expected 2, got %d", s.ZCard("z"))
	}
}

// --- Wrong type ---

func TestZSet_WrongType(t *testing.T) {
	s := NewStore()
	s.Set("key", "string")
	_, err := s.ZAdd("key", "10", "x")
	if err == nil {
		t.Error("ZAdd on string key should return error")
	}
}

// --- 并发 ---

func TestZSet_Concurrent(t *testing.T) {
	s := NewStore()
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			s.ZAdd("z", "1", "a")
			s.ZScore("z", "a")
			s.ZRank("z", "a")
			s.ZRange("z", 0, -1)
			s.ZCard("z")
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
