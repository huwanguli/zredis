package core

import "testing"

// --- StringVal ---

func TestStringVal_Type(t *testing.T) {
	v := NewStringVal("OK")
	if v.Type() != TypeString {
		t.Errorf("expected TypeString, got %c", v.Type())
	}
}

func TestStringVal_String(t *testing.T) {
	v := NewStringVal("OK")
	if v.String() != `StringVal("OK")` {
		t.Errorf("unexpected String(): %s", v.String())
	}
}

// --- ErrorVal ---

func TestErrorVal_Type(t *testing.T) {
	v := NewErrorVal("ERR unknown command")
	if v.Type() != TypeError {
		t.Errorf("expected TypeError, got %c", v.Type())
	}
}

func TestErrorVal_String(t *testing.T) {
	v := NewErrorVal("ERR something went wrong")
	if v.String() != `ErrorVal("ERR something went wrong")` {
		t.Errorf("unexpected String(): %s", v.String())
	}
}

// --- IntVal ---

func TestIntVal_Type(t *testing.T) {
	v := NewIntVal(42)
	if v.Type() != TypeInteger {
		t.Errorf("expected TypeInteger, got %c", v.Type())
	}
}

func TestIntVal_String(t *testing.T) {
	v := NewIntVal(-7)
	if v.String() != "IntVal(-7)" {
		t.Errorf("unexpected String(): %s", v.String())
	}
}

func TestIntVal_Zero(t *testing.T) {
	v := NewIntVal(0)
	if v.N != 0 {
		t.Errorf("expected 0, got %d", v.N)
	}
}

// --- BulkVal ---

func TestBulkVal_Type(t *testing.T) {
	v := NewBulkVal([]byte("hello"))
	if v.Type() != TypeBulk {
		t.Errorf("expected TypeBulk, got %c", v.Type())
	}
}

func TestBulkVal_String_Normal(t *testing.T) {
	v := NewBulkVal([]byte("hello"))
	if v.String() != `BulkVal("hello")` {
		t.Errorf("unexpected String(): %s", v.String())
	}
}

func TestBulkVal_String_Null(t *testing.T) {
	v := NewNullBulkVal()
	if v.String() != "BulkVal(null)" {
		t.Errorf("unexpected String(): %s", v.String())
	}
}

func TestBulkVal_NullIsDistinct(t *testing.T) {
	nullVal := NewNullBulkVal()
	emptyVal := NewBulkVal([]byte{})

	if nullVal.Data != nil {
		t.Error("null bulk val should have nil Data")
	}
	if emptyVal.Data == nil {
		t.Error("empty bulk val should have non-nil Data")
	}
}

func TestBulkVal_BinarySafe(t *testing.T) {
	data := []byte{0x00, 0xFF, 0x0D, 0x0A, 0x7F}
	v := NewBulkVal(data)
	if len(v.Data) != len(data) {
		t.Errorf("expected length %d, got %d", len(data), len(v.Data))
	}
	for i, b := range data {
		if v.Data[i] != b {
			t.Errorf("byte mismatch at index %d: expected %02x, got %02x", i, b, v.Data[i])
		}
	}
}

// --- ArrayVal ---

func TestArrayVal_Type(t *testing.T) {
	v := NewArrayVal([]Value{NewStringVal("foo"), NewIntVal(3)})
	if v.Type() != TypeArray {
		t.Errorf("expected TypeArray, got %c", v.Type())
	}
}

func TestArrayVal_String_Normal(t *testing.T) {
	v := NewArrayVal([]Value{NewStringVal("a")})
	expected := "ArrayVal(len=1)"
	if v.String() != expected {
		t.Errorf("expected %q, got %q", expected, v.String())
	}
}

func TestArrayVal_String_Empty(t *testing.T) {
	v := NewArrayVal([]Value{})
	if v.String() != "ArrayVal(len=0)" {
		t.Errorf("unexpected String(): %s", v.String())
	}
}

func TestArrayVal_String_Null(t *testing.T) {
	v := NewNullArrayVal()
	if v.String() != "ArrayVal(null)" {
		t.Errorf("unexpected String(): %s", v.String())
	}
}

func TestArrayVal_NullIsDistinct(t *testing.T) {
	nullArr := NewNullArrayVal()
	emptyArr := NewArrayVal([]Value{})

	if nullArr.Items != nil {
		t.Error("null array should have nil Items")
	}
	if emptyArr.Items == nil {
		t.Error("empty array should have non-nil Items")
	}
}

// --- Interface uses ---

func TestValuesAsInterface(t *testing.T) {
	vals := []Value{
		NewStringVal("OK"),
		NewErrorVal("ERR fail"),
		NewIntVal(10),
		NewBulkVal([]byte("data")),
		NewNullBulkVal(),
		NewArrayVal([]Value{NewStringVal("nested")}),
		NewNullArrayVal(),
	}

	expectedTypes := []ValueType{TypeString, TypeError, TypeInteger, TypeBulk, TypeBulk, TypeArray, TypeArray}
	for i, v := range vals {
		if v.Type() != expectedTypes[i] {
			t.Errorf("vals[%d]: expected type %c, got %c", i, expectedTypes[i], v.Type())
		}
	}
}

// --- Compile-time interface satisfaction ---

func TestCompileTimeInterfaceCheck(t *testing.T) {
	// This test exists to ensure the _ = Value(...) lines in value.go compile.
	// If they compile, this test passes. No runtime checks needed.
}
