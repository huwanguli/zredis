package core

import (
	"testing"
)

func dispatch(s *Store, args ...string) Value {
	d := NewDispatcher()
	items := make([]Value, len(args))
	for i, a := range args {
		items[i] = NewBulkVal([]byte(a))
	}
	return d.Dispatch(s, NewArrayVal(items))
}

// --- 通用 ---

func TestDispatch_Ping(t *testing.T) {
	result := dispatch(NewStore(), "PING")
	sv, ok := result.(*StringVal)
	if !ok || sv.Str != "PONG" {
		t.Errorf("expected PONG, got %v", result)
	}
}

func TestDispatch_Del(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("b", "2")
	result := dispatch(s, "DEL", "a", "b", "c")
	iv, ok := result.(*IntVal)
	if !ok || iv.N != 2 {
		t.Errorf("DEL expected 2, got %v", result)
	}
}

func TestDispatch_Exists(t *testing.T) {
	s := NewStore()
	s.Set("x", "1")
	r1 := dispatch(s, "EXISTS", "x")
	r2 := dispatch(s, "EXISTS", "y")
	if r1.(*IntVal).N != 1 || r2.(*IntVal).N != 0 {
		t.Errorf("EXISTS failed: %v %v", r1, r2)
	}
}

func TestDispatch_Keys(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("b", "2")
	result := dispatch(s, "KEYS")
	av := result.(*ArrayVal)
	if len(av.Items) != 2 {
		t.Errorf("KEYS expected 2, got %d", len(av.Items))
	}
}

func TestDispatch_Expire_TTL(t *testing.T) {
	s := NewStore()
	s.Set("x", "1")
	r1 := dispatch(s, "EXPIRE", "x", "100")
	r2 := dispatch(s, "TTL", "x")
	r3 := dispatch(s, "TTL", "no")
	if r1.(*IntVal).N != 1 {
		t.Error("EXPIRE failed")
	}
	if r2.(*IntVal).N <= 0 {
		t.Error("TTL should be positive")
	}
	if r3.(*IntVal).N != -2 {
		t.Errorf("TTL for no key expected -2, got %d", r3.(*IntVal).N)
	}
}

// --- String ---

func TestDispatch_SetGet(t *testing.T) {
	s := NewStore()
	dispatch(s, "SET", "k", "hello")
	result := dispatch(s, "GET", "k")
	bv := result.(*BulkVal)
	if string(bv.Data) != "hello" {
		t.Errorf("GET failed: %v", result)
	}
}

func TestDispatch_IncrDecr(t *testing.T) {
	s := NewStore()
	s.Set("c", "10")
	r1 := dispatch(s, "INCR", "c")
	r2 := dispatch(s, "DECR", "c")
	if r1.(*IntVal).N != 11 || r2.(*IntVal).N != 10 {
		t.Errorf("INCR/DECR failed: %v %v", r1, r2)
	}
}

func TestDispatch_MGetMSet(t *testing.T) {
	s := NewStore()
	dispatch(s, "MSET", "a", "1", "b", "2")
	r := dispatch(s, "MGET", "a", "b", "c")
	av := r.(*ArrayVal)
	if len(av.Items) != 3 {
		t.Fatalf("MGET len expected 3, got %d", len(av.Items))
	}
	if string(av.Items[0].(*BulkVal).Data) != "1" {
		t.Error("MGET[0] wrong")
	}
	if string(av.Items[1].(*BulkVal).Data) != "2" {
		t.Error("MGET[1] wrong")
	}
	if av.Items[2].(*BulkVal).Data != nil {
		t.Error("MGET[2] should be null")
	}
}

// --- Hash ---

func TestDispatch_Hash(t *testing.T) {
	s := NewStore()
	r1 := dispatch(s, "HSET", "h", "f1", "v1", "f2", "v2")
	r2 := dispatch(s, "HGET", "h", "f1")
	r3 := dispatch(s, "HLEN", "h")
	r4 := dispatch(s, "HEXISTS", "h", "f2")
	r5 := dispatch(s, "HEXISTS", "h", "f3")

	if r1.(*IntVal).N != 2 {
		t.Error("HSET failed")
	}
	if string(r2.(*BulkVal).Data) != "v1" {
		t.Error("HGET failed")
	}
	if r3.(*IntVal).N != 2 {
		t.Error("HLEN failed")
	}
	if r4.(*IntVal).N != 1 {
		t.Error("HEXISTS failed")
	}
	if r5.(*IntVal).N != 0 {
		t.Error("HEXISTS should be 0")
	}
}

func TestDispatch_HGetAll(t *testing.T) {
	s := NewStore()
	dispatch(s, "HSET", "h", "a", "1", "b", "2")
	r := dispatch(s, "HGETALL", "h")
	av := r.(*ArrayVal)
	if len(av.Items) != 4 {
		t.Fatalf("HGETALL expected 4 items, got %d", len(av.Items))
	}
}

// --- List ---

func TestDispatch_List(t *testing.T) {
	s := NewStore()
	r1 := dispatch(s, "LPUSH", "l", "b", "a")
	r2 := dispatch(s, "RPUSH", "l", "c", "d")
	r3 := dispatch(s, "LLEN", "l")

	if r1.(*IntVal).N != 2 {
		t.Error("LPUSH failed")
	}
	if r2.(*IntVal).N != 4 {
		t.Error("RPUSH failed")
	}
	if r3.(*IntVal).N != 4 {
		t.Error("LLEN failed")
	}
}

func TestDispatch_LRange(t *testing.T) {
	s := NewStore()
	dispatch(s, "RPUSH", "l", "a", "b", "c")
	r := dispatch(s, "LRANGE", "l", "0", "-1")
	av := r.(*ArrayVal)
	if len(av.Items) != 3 {
		t.Fatalf("LRANGE expected 3, got %d", len(av.Items))
	}
}

func TestDispatch_LPopRPop(t *testing.T) {
	s := NewStore()
	dispatch(s, "RPUSH", "l", "x", "y", "z")
	r1 := dispatch(s, "LPOP", "l")
	r2 := dispatch(s, "RPOP", "l")
	if string(r1.(*BulkVal).Data) != "x" {
		t.Error("LPOP failed")
	}
	if string(r2.(*BulkVal).Data) != "z" {
		t.Error("RPOP failed")
	}
}

// --- Set ---

func TestDispatch_Set(t *testing.T) {
	s := NewStore()
	r1 := dispatch(s, "SADD", "s", "a", "b", "c")
	r2 := dispatch(s, "SADD", "s", "c", "d")
	r3 := dispatch(s, "SCARD", "s")
	r4 := dispatch(s, "SISMEMBER", "s", "a")
	r5 := dispatch(s, "SISMEMBER", "s", "x")

	if r1.(*IntVal).N != 3 {
		t.Error("SADD failed")
	}
	if r2.(*IntVal).N != 1 {
		t.Error("SADD second failed")
	}
	if r3.(*IntVal).N != 4 {
		t.Error("SCARD failed")
	}
	if r4.(*IntVal).N != 1 {
		t.Error("SISMEMBER failed")
	}
	if r5.(*IntVal).N != 0 {
		t.Error("SISMEMBER should be 0")
	}
}

func TestDispatch_SetOps(t *testing.T) {
	s := NewStore()
	dispatch(s, "SADD", "s1", "a", "b", "c")
	dispatch(s, "SADD", "s2", "b", "c", "d")

	inter := dispatch(s, "SINTER", "s1", "s2")
	union := dispatch(s, "SUNION", "s1", "s2")
	diff := dispatch(s, "SDIFF", "s1", "s2")

	if len(inter.(*ArrayVal).Items) != 2 {
		t.Error("SINTER failed")
	}
	if len(union.(*ArrayVal).Items) != 4 {
		t.Error("SUNION failed")
	}
	if len(diff.(*ArrayVal).Items) != 1 {
		t.Error("SDIFF failed")
	}
}

// --- ZSet ---

func TestDispatch_ZSet(t *testing.T) {
	s := NewStore()
	r1 := dispatch(s, "ZADD", "z", "10", "a", "20", "b")
	r2 := dispatch(s, "ZCARD", "z")
	r3 := dispatch(s, "ZRANK", "z", "b")
	r4 := dispatch(s, "ZSCORE", "z", "a")

	if r1.(*IntVal).N != 2 {
		t.Error("ZADD failed")
	}
	if r2.(*IntVal).N != 2 {
		t.Error("ZCARD failed")
	}
	if r3.(*IntVal).N != 1 {
		t.Errorf("ZRANK expected 1, got %d", r3.(*IntVal).N)
	}
	bv := r4.(*BulkVal)
	if bv.Data == nil || string(bv.Data) != "10" {
		t.Error("ZSCORE failed")
	}
}

// --- 边界 ---

func TestDispatch_UnknownCommand(t *testing.T) {
	result := dispatch(NewStore(), "FOOBAR")
	ev, ok := result.(*ErrorVal)
	if !ok {
		t.Fatalf("expected ErrorVal, got %T", result)
	}
	if !stringsContains(ev.Msg, "unknown") {
		t.Errorf("unexpected error msg: %s", ev.Msg)
	}
}

func TestDispatch_EmptyCommand(t *testing.T) {
	d := NewDispatcher()
	result := d.Dispatch(NewStore(), NewArrayVal([]Value{}))
	if _, ok := result.(*ErrorVal); !ok {
		t.Errorf("expected error for empty command, got %v", result)
	}
}

func TestDispatch_WrongType(t *testing.T) {
	s := NewStore()
	s.Set("x", "hello")
	result := dispatch(s, "LPUSH", "x", "a")
	if _, ok := result.(*ErrorVal); !ok {
		t.Errorf("expected error for wrong type, got %v", result)
	}
}

func stringsContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
