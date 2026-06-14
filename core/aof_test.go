package core

import (
	"os"
	"testing"
)

func TestAOF_WriteAndLoad(t *testing.T) {
	path := "test_aof.log"
	os.Remove(path) // 清理上次残留
	defer os.Remove(path)

	aof, err := NewAOF(path)
	if err != nil {
		t.Fatalf("NewAOF: %v", err)
	}

	// 写入几条命令
	cmds := []*ArrayVal{
		NewArrayVal([]Value{NewBulkVal([]byte("SET")), NewBulkVal([]byte("k1")), NewBulkVal([]byte("v1"))}),
		NewArrayVal([]Value{NewBulkVal([]byte("SET")), NewBulkVal([]byte("k2")), NewBulkVal([]byte("v2"))}),
		NewArrayVal([]Value{NewBulkVal([]byte("INCR")), NewBulkVal([]byte("count"))}),
		NewArrayVal([]Value{NewBulkVal([]byte("INCR")), NewBulkVal([]byte("count"))}),
	}
	for _, cmd := range cmds {
		if err := aof.Write(cmd); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	// 重放到新的 Store
	s := NewStore()
	d := NewDispatcher()
	if err := aof.Load(s, d); err != nil {
		t.Fatalf("Load: %v", err)
	}
	aof.Close()

	// 验证
	v1, ok := s.Get("k1")
	if !ok || v1 != "v1" {
		t.Errorf("expected v1, got %q (ok=%v)", v1, ok)
	}
	v2, ok := s.Get("k2")
	if !ok || v2 != "v2" {
		t.Errorf("expected v2, got %q (ok=%v)", v2, ok)
	}
	n, ok := s.Get("count")
	if !ok || n != "2" {
		t.Errorf("expected count=2, got %q (ok=%v)", n, ok)
	}
}

func TestAOF_EmptyFile(t *testing.T) {
	path := "test_empty.log"
	os.Remove(path)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(path)

	aof, err := NewAOF(path)
	if err != nil {
		t.Fatalf("NewAOF: %v", err)
	}
	defer aof.Close()

	s := NewStore()
	d := NewDispatcher()
	if err := aof.Load(s, d); err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("expected empty store, got %d keys", s.Len())
	}
}

func TestIsWriteCmd(t *testing.T) {
	writes := []string{"SET", "DEL", "INCR", "HSET", "LPUSH", "SADD", "ZADD"}
	reads := []string{"GET", "KEYS", "PING", "HGET", "LRANGE", "SMEMBERS", "ZRANGE"}

	for _, c := range writes {
		if !IsWriteCmd(c) {
			t.Errorf("%s should be a write command", c)
		}
	}
	for _, c := range reads {
		if IsWriteCmd(c) {
			t.Errorf("%s should NOT be a write command", c)
		}
	}
}
