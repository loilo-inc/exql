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

func TestNewTypedNull(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		n := New[int64](123)
		if !n.Valid || n.V != 123 {
			t.Errorf("New[int64](123) = %v, want valid=true V=123", n)
		}
	})

	t.Run("string", func(t *testing.T) {
		n := New("test")
		if !n.Valid || n.V != "test" {
			t.Errorf("New(\"test\") = %v, want valid=true V=test", n)
		}
	})

	t.Run("float64", func(t *testing.T) {
		n := New(3.14)
		if !n.Valid || n.V != 3.14 {
			t.Errorf("New(3.14) = %v, want valid=true V=3.14", n)
		}
	})
}

func TestFromPtr(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		v := int64(123)
		n := FromPtr(&v)
		if !n.Valid || n.V != 123 {
			t.Errorf("FromPtr(&%v) = %v, want valid=true V=123", v, n)
		}
	})

	t.Run("string", func(t *testing.T) {
		v := "test"
		n := FromPtr(&v)
		if !n.Valid || n.V != "test" {
			t.Errorf("FromPtr(&%q) = %v, want valid=true V=test", v, n)
		}
	})

	t.Run("float64", func(t *testing.T) {
		v := 3.14
		n := FromPtr(&v)
		if !n.Valid || n.V != 3.14 {
			t.Errorf("FromPtr(&%v) = %v, want valid=true V=3.14", v, n)
		}
	})

	t.Run("nil", func(t *testing.T) {
		n := FromPtr[int](nil)
		if n.Valid {
			t.Errorf("FromPtr(nil) = %v, want valid=false", n)
		}
	})
}

func TestNullPtr(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		n := New(42)
		ptr := n.Ptr()
		if ptr == nil || *ptr != 42 {
			t.Errorf("Ptr() = %v, want pointer to 42", ptr)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		n := Null[int]{}
		ptr := n.Ptr()
		if ptr != nil {
			t.Errorf("Ptr() on invalid Null = %v, want nil", ptr)
		}
	})
	t.Run("string", func(t *testing.T) {
		n := New("test")
		ptr := n.Ptr()
		if ptr == nil || *ptr != "test" {
			t.Errorf("Ptr() = %v, want pointer to \"test\"", ptr)
		}
	})
	t.Run("float64", func(t *testing.T) {
		n := New(2.71)
		ptr := n.Ptr()
		if ptr == nil || *ptr != 2.71 {
			t.Errorf("Ptr() = %v, want pointer to 2.71", ptr)
		}
	})
}
