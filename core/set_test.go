package core

import "testing"

// --- SAdd / SRem ---

func TestSAdd_Basic(t *testing.T) {
	s := NewStore()
	n, err := s.SAdd("set", "a", "b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}

func TestSAdd_Duplicate(t *testing.T) {
	s := NewStore()
	s.SAdd("set", "a", "b")
	n, _ := s.SAdd("set", "b", "c")
	if n != 1 {
		t.Errorf("expected 1 new member, got %d", n)
	}
}

func TestSRem_Basic(t *testing.T) {
	s := NewStore()
	s.SAdd("set", "a", "b", "c")
	n, _ := s.SRem("set", "a", "c", "d")
	if n != 2 {
		t.Errorf("expected 2 removed, got %d", n)
	}
	if s.SIsMember("set", "a") {
		t.Error("a should be removed")
	}
	if !s.SIsMember("set", "b") {
		t.Error("b should still exist")
	}
}

// --- SMembers ---

func TestSMembers_Basic(t *testing.T) {
	s := NewStore()
	s.SAdd("set", "a", "b")
	members := s.SMembers("set")
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
}

func TestSMembers_NonExistent(t *testing.T) {
	s := NewStore()
	if members := s.SMembers("noexist"); members != nil {
		t.Error("SMembers should return nil for non-existent key")
	}
}

// --- SIsMember ---

func TestSIsMember_Basic(t *testing.T) {
	s := NewStore()
	s.SAdd("set", "a")
	if !s.SIsMember("set", "a") {
		t.Error("should be member")
	}
	if s.SIsMember("set", "b") {
		t.Error("should not be member")
	}
}

// --- SCard ---

func TestSCard_Basic(t *testing.T) {
	s := NewStore()
	if s.SCard("set") != 0 {
		t.Error("SCard should return 0 for non-existent key")
	}
	s.SAdd("set", "a", "b")
	if s.SCard("set") != 2 {
		t.Errorf("expected 2, got %d", s.SCard("set"))
	}
}

// --- SInter ---

func TestSInter_Basic(t *testing.T) {
	s := NewStore()
	s.SAdd("s1", "a", "b", "c")
	s.SAdd("s2", "b", "c", "d")

	result := s.SInter("s1", "s2")
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

// --- SUnion ---

func TestSUnion_Basic(t *testing.T) {
	s := NewStore()
	s.SAdd("s1", "a", "b")
	s.SAdd("s2", "b", "c")

	result := s.SUnion("s1", "s2")
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
}

// --- SDiff ---

func TestSDiff_Basic(t *testing.T) {
	s := NewStore()
	s.SAdd("s1", "a", "b", "c")
	s.SAdd("s2", "b")

	result := s.SDiff("s1", "s2")
	if len(result) != 2 {
		t.Errorf("expected 2 (a, c), got %d", len(result))
	}
}

// --- Wrong type ---

func TestSet_WrongType(t *testing.T) {
	s := NewStore()
	s.Set("key", "string")
	_, err := s.SAdd("key", "x")
	if err == nil {
		t.Error("SAdd on string key should return error")
	}
}

// --- 并发 ---

func TestSet_Concurrent(t *testing.T) {
	s := NewStore()
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			s.SAdd("set", "x")
			s.SRem("set", "x")
			s.SIsMember("set", "x")
			s.SCard("set")
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
