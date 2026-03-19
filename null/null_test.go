package null

import (
	"database/sql"
	"testing"
)

func TestNewNull(t *testing.T) {
	n := New(42)
	if !n.Valid || n.V != 42 {
		t.Errorf("New(42) = %v, want valid=true V=42", n)
	}
}

func TestNullMarshalJSONValid(t *testing.T) {
	n := New("hello")
	data, err := n.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(data) != `"hello"` {
		t.Errorf("MarshalJSON() = %s, want \"hello\"", data)
	}
}

func TestNullMarshalJSONInvalid(t *testing.T) {
	n := Null[string]{sql.Null[string]{Valid: false}}
	data, err := n.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON() = %s, want null", data)
	}
}

func TestNullInt64(t *testing.T) {
	n := New[int64](123)
	if !n.Valid || n.V != 123 {
		t.Errorf("Int64 type = %v, want valid=true V=123", n)
	}
}

func TestNullString(t *testing.T) {
	n := New("test")
	if !n.Valid || n.V != "test" {
		t.Errorf("String type = %v, want valid=true V=test", n)
	}
}

func TestNullFloat64(t *testing.T) {
	n := New(3.14)
	if !n.Valid || n.V != 3.14 {
		t.Errorf("Float64 type = %v, want valid=true V=3.14", n)
	}
}
