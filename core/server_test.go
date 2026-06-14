package core

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Remove("appendonly.aof") // 清理上次残留
	code := m.Run()
	os.Remove("appendonly.aof")
	os.Exit(code)
}

func TestServer_Ping(t *testing.T) {
	srv := NewServer()

	addr := "127.0.0.1:16380"
	go srv.Listen(addr)
	time.Sleep(50 * time.Millisecond) // 等 server 启动

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	r := NewReader(conn)
	w := NewWriter(conn)

	// PING
	w.WriteValue(NewArrayVal([]Value{NewBulkVal([]byte("PING"))}))
	w.Flush()
	v, err := r.ReadValue()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	sv, ok := v.(*StringVal)
	if !ok || sv.Str != "PONG" {
		t.Errorf("expected PONG, got %v", v)
	}
}

func TestServer_SetGet(t *testing.T) {
	srv := NewServer()

	addr := "127.0.0.1:16381"
	go srv.Listen(addr)
	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	r := NewReader(conn)
	w := NewWriter(conn)

	// SET mykey hello
	w.WriteValue(NewArrayVal([]Value{
		NewBulkVal([]byte("SET")),
		NewBulkVal([]byte("mykey")),
		NewBulkVal([]byte("hello")),
	}))
	w.Flush()
	v, _ := r.ReadValue()
	if sv, ok := v.(*StringVal); !ok || sv.Str != "OK" {
		t.Errorf("SET failed: %v", v)
	}

	// GET mykey
	w.WriteValue(NewArrayVal([]Value{
		NewBulkVal([]byte("GET")),
		NewBulkVal([]byte("mykey")),
	}))
	w.Flush()
	v, _ = r.ReadValue()
	bv, ok := v.(*BulkVal)
	if !ok || string(bv.Data) != "hello" {
		t.Errorf("GET failed: %v", v)
	}
}

func TestServer_MultipleCommands(t *testing.T) {
	srv := NewServer()

	addr := "127.0.0.1:16382"
	go srv.Listen(addr)
	time.Sleep(50 * time.Millisecond)

	conn, _ := net.Dial("tcp", addr)
	defer conn.Close()
	r := NewReader(conn)
	w := NewWriter(conn)

	cmds := [][]Value{
		{NewBulkVal([]byte("SET")), NewBulkVal([]byte("x")), NewBulkVal([]byte("1"))},
		{NewBulkVal([]byte("INCR")), NewBulkVal([]byte("x"))},
		{NewBulkVal([]byte("INCR")), NewBulkVal([]byte("x"))},
		{NewBulkVal([]byte("GET")), NewBulkVal([]byte("x"))},
	}

	expected := []string{"OK", "2", "3", "3"}
	for i, args := range cmds {
		w.WriteValue(NewArrayVal(args))
		w.Flush()
		v, err := r.ReadValue()
		if err != nil {
			t.Fatalf("cmd %d read error: %v", i, err)
		}
		got := fmt.Sprint(v)
		if !contains(got, expected[i]) {
			t.Errorf("cmd %d: expected containing %q, got %s", i, expected[i], got)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchSub(s, sub)
}

func searchSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
