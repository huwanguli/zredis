package core

import (
	"bytes"
	"testing"
)

func TestWriteSimpleString(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	err := w.WriteValue(NewStringVal("OK"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w.Flush()
	if buf.String() != "+OK\r\n" {
		t.Errorf("expected '+OK\\r\\n', got '%s'", buf.String())
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewErrorVal("ERR something"))
	w.Flush()
	if buf.String() != "-ERR something\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestWriteInteger(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewIntVal(42))
	w.Flush()
	if buf.String() != ":42\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestWriteInteger_Negative(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewIntVal(-7))
	w.Flush()
	if buf.String() != ":-7\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestWriteBulkString(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewBulkVal([]byte("hello")))
	w.Flush()
	if buf.String() != "$5\r\nhello\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestWriteBulkString_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewBulkVal([]byte{}))
	w.Flush()
	if buf.String() != "$0\r\n\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestWriteBulkString_Null(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewNullBulkVal())
	w.Flush()
	if buf.String() != "$-1\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestWriteArray(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewArrayVal([]Value{
		NewBulkVal([]byte("foo")),
		NewBulkVal([]byte("bar")),
	}))
	w.Flush()
	expected := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	if buf.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, buf.String())
	}
}

func TestWriteArray_Null(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewNullArrayVal())
	w.Flush()
	if buf.String() != "*-1\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestWriteArray_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteValue(NewArrayVal([]Value{}))
	w.Flush()
	if buf.String() != "*0\r\n" {
		t.Errorf("got '%s'", buf.String())
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"+OK\r\n",
		"-ERR fail\r\n",
		":1234\r\n",
		"$5\r\nhello\r\n",
		"$0\r\n\r\n",
		"$-1\r\n",
		"*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
	}

	for _, input := range inputs {
		r := NewReader(bytes.NewReader([]byte(input)))
		v, err := r.ReadValue()
		if err != nil {
			t.Errorf("read error for '%s': %v", input, err)
			continue
		}

		var buf bytes.Buffer
		w := NewWriter(&buf)
		if err := w.WriteValue(v); err != nil {
			t.Errorf("write error for '%s': %v", input, err)
			continue
		}
		w.Flush()

		if buf.String() != input {
			t.Errorf("round-trip mismatch for '%s': got '%s'", input, buf.String())
		}
	}
}
