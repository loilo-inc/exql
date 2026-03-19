package null

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type unsupportedTextValue struct {
	Name string
}

type namedString string
type namedInt int

type textJSONPayload struct {
	A int `json:"a"`
}

func TestNewNull(t *testing.T) {
	n := New(42)
	assert.True(t, n.Valid)
	assert.Equal(t, 42, n.V)
}

func TestNullMarshalJSONValid(t *testing.T) {
	n := New("hello")
	data, err := n.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `"hello"`, string(data))
}

func TestNullMarshalJSONInvalid(t *testing.T) {
	n := Null[string]{sql.Null[string]{Valid: false}}
	data, err := n.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

func TestNullUnmarshalJSON(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		var n Null[string]
		err := n.UnmarshalJSON([]byte(`"hello"`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, "hello", n.V)
	})

	t.Run("int", func(t *testing.T) {
		var n Null[int]
		err := n.UnmarshalJSON([]byte(`123`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, 123, n.V)
	})

	t.Run("int64", func(t *testing.T) {
		var n Null[int64]
		err := n.UnmarshalJSON([]byte(`1234567890123`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.EqualValues(t, 1234567890123, n.V)
	})

	t.Run("float", func(t *testing.T) {
		var n Null[float64]
		err := n.UnmarshalJSON([]byte(`3.14`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, 3.14, n.V)
	})

	t.Run("null", func(t *testing.T) {
		n := New("hello")
		err := n.UnmarshalJSON([]byte(`null`))
		assert.NoError(t, err)
		assert.False(t, n.Valid)
		assert.Equal(t, "", n.V)
	})

	t.Run("invalid json", func(t *testing.T) {
		n := New("before")
		err := n.UnmarshalJSON([]byte(`invalid`))
		assert.Error(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, "before", n.V)
	})

	t.Run("time valid value", func(t *testing.T) {
		var n Time
		want := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
		err := n.UnmarshalJSON([]byte(`"2024-01-02T03:04:05Z"`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.True(t, n.V.Equal(want))
	})

	t.Run("time null", func(t *testing.T) {
		n := New(time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC))
		err := n.UnmarshalJSON([]byte(`null`))
		assert.NoError(t, err)
		assert.False(t, n.Valid)
		assert.True(t, n.V.IsZero())
	})

	t.Run("time null with whitespace", func(t *testing.T) {
		n := New(time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC))
		err := n.UnmarshalJSON([]byte(" \n\t null\r\n "))
		assert.NoError(t, err)
		assert.False(t, n.Valid)
		assert.True(t, n.V.IsZero())
	})

	t.Run("bytes valid value", func(t *testing.T) {
		var n Bytes
		err := n.UnmarshalJSON([]byte(`"aGVsbG8="`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, []byte("hello"), n.V)
	})

	t.Run("bytes null", func(t *testing.T) {
		n := New([]byte("hello"))
		err := n.UnmarshalJSON([]byte(`null`))
		assert.NoError(t, err)
		assert.False(t, n.Valid)
		assert.Nil(t, n.V)
	})
}

func TestNullUnmarshalText(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		var n Null[string]
		err := n.UnmarshalText([]byte(`hello`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, "hello", n.V)
	})

	t.Run("quoted text is preserved", func(t *testing.T) {
		var n Null[string]
		err := n.UnmarshalText([]byte(`"hello"`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, `"hello"`, n.V)
	})

	t.Run("integer", func(t *testing.T) {
		var n Null[int]
		err := n.UnmarshalText([]byte(`123`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, 123, n.V)
	})

	t.Run("named string", func(t *testing.T) {
		var n Null[namedString]
		err := n.UnmarshalText([]byte(`hello`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, namedString("hello"), n.V)
	})

	t.Run("named integer", func(t *testing.T) {
		var n Null[namedInt]
		err := n.UnmarshalText([]byte(`123`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, namedInt(123), n.V)
	})

	t.Run("float", func(t *testing.T) {
		var n Null[float64]
		err := n.UnmarshalText([]byte(`3.14`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, 3.14, n.V)
	})

	t.Run("empty text", func(t *testing.T) {
		n := New(time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC))
		err := n.UnmarshalText([]byte(""))
		assert.NoError(t, err)
		assert.False(t, n.Valid)
		assert.True(t, n.V.IsZero())
	})

	t.Run("whitespace text", func(t *testing.T) {
		n := New(time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC))
		err := n.UnmarshalText([]byte(" \n\t "))
		assert.Error(t, err)
		assert.True(t, n.Valid)
	})

	t.Run("plain string text is accepted", func(t *testing.T) {
		n := New("before")
		err := n.UnmarshalText([]byte(`invalid`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, "invalid", n.V)
	})

	t.Run("time valid value", func(t *testing.T) {
		var n Time
		want := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
		err := n.UnmarshalText([]byte(`2024-01-02T03:04:05Z`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.True(t, n.V.Equal(want))
	})

	t.Run("time valid JSON-encoded value", func(t *testing.T) {
		var n Time
		err := n.UnmarshalText([]byte(`"2024-01-02T03:04:05Z"`))
		assert.Error(t, err)
		assert.False(t, n.Valid)
		assert.True(t, n.V.IsZero())
	})

	t.Run("struct valid JSON value", func(t *testing.T) {
		var n Null[textJSONPayload]
		err := n.UnmarshalText([]byte(`{"a":1}`))
		assert.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, textJSONPayload{A: 1}, n.V)
	})

	t.Run("unsupported type", func(t *testing.T) {
		n := New(unsupportedTextValue{Name: "before"})
		err := n.UnmarshalText([]byte(`anything`))
		assert.ErrorIs(t, err, errUnsupportedType)
		assert.True(t, n.Valid)
		assert.Equal(t, unsupportedTextValue{Name: "before"}, n.V)
	})
}

func TestNewTypedNull(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		n := New[int64](123)
		assert.True(t, n.Valid)
		assert.EqualValues(t, 123, n.V)
	})

	t.Run("string", func(t *testing.T) {
		n := New("test")
		assert.True(t, n.Valid)
		assert.Equal(t, "test", n.V)
	})

	t.Run("float64", func(t *testing.T) {
		n := New(3.14)
		assert.True(t, n.Valid)
		assert.Equal(t, 3.14, n.V)
	})
}

func TestFromPtr(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		v := int64(123)
		n := FromPtr(&v)
		assert.True(t, n.Valid)
		assert.EqualValues(t, 123, n.V)
	})

	t.Run("string", func(t *testing.T) {
		v := "test"
		n := FromPtr(&v)
		assert.True(t, n.Valid)
		assert.Equal(t, "test", n.V)
	})

	t.Run("float64", func(t *testing.T) {
		v := 3.14
		n := FromPtr(&v)
		assert.True(t, n.Valid)
		assert.Equal(t, 3.14, n.V)
	})

	t.Run("nil", func(t *testing.T) {
		n := FromPtr[int](nil)
		assert.False(t, n.Valid)
	})
}

func TestNullPtr(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		n := New(42)
		ptr := n.Ptr()
		assert.NotNil(t, ptr)
		assert.Equal(t, 42, *ptr)
	})
	t.Run("invalid", func(t *testing.T) {
		n := Null[int]{}
		ptr := n.Ptr()
		assert.Nil(t, ptr)
	})
	t.Run("string", func(t *testing.T) {
		n := New("test")
		ptr := n.Ptr()
		assert.NotNil(t, ptr)
		assert.Equal(t, "test", *ptr)
	})
	t.Run("float64", func(t *testing.T) {
		n := New(2.71)
		ptr := n.Ptr()
		assert.NotNil(t, ptr)
		assert.Equal(t, 2.71, *ptr)
	})
}
