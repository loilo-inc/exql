package null_test

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	exqlnull "github.com/loilo-inc/exql/v3/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compatibility tests to volatiletech/null

func TestCompatibilityConstructors(t *testing.T) {
	t.Run("new and ptr round trip", func(t *testing.T) {
		v := "test"
		n := exqlnull.New(v)
		require.True(t, n.Valid)
		assert.Equal(t, "test", n.V)
		require.NotNil(t, n.Ptr())
		assert.Equal(t, "test", *n.Ptr())
	})

	t.Run("from nil ptr stays invalid", func(t *testing.T) {
		n := exqlnull.FromPtr[int](nil)
		assert.False(t, n.Valid)
		assert.Nil(t, n.Ptr())
	})

	t.Run("zero value can still be valid", func(t *testing.T) {
		n := exqlnull.New(false)
		assert.True(t, n.Valid)
		assert.False(t, n.V)
	})
}

func TestCompatibilityJSON(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		n := exqlnull.New("test")
		data, err := json.Marshal(n)
		require.NoError(t, err)
		assert.JSONEq(t, `"test"`, string(data))

		var decoded exqlnull.String
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.True(t, decoded.Valid)
		assert.Equal(t, "test", decoded.V)
	})

	t.Run("bool", func(t *testing.T) {
		n := exqlnull.New(true)
		data, err := json.Marshal(n)
		require.NoError(t, err)
		assert.JSONEq(t, `true`, string(data))

		var decoded exqlnull.Bool
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.True(t, decoded.Valid)
		assert.True(t, decoded.V)
	})

	t.Run("int", func(t *testing.T) {
		n := exqlnull.New(12345)
		data, err := json.Marshal(n)
		require.NoError(t, err)
		assert.JSONEq(t, `12345`, string(data))

		var decoded exqlnull.Null[int]
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.True(t, decoded.Valid)
		assert.Equal(t, 12345, decoded.V)
	})

	t.Run("bytes", func(t *testing.T) {
		n := exqlnull.New([]byte("hello"))
		data, err := json.Marshal(n)
		require.NoError(t, err)
		assert.JSONEq(t, `"aGVsbG8="`, string(data))

		var decoded exqlnull.Bytes
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.True(t, decoded.Valid)
		assert.Equal(t, []byte("hello"), decoded.V)
	})

	t.Run("time", func(t *testing.T) {
		want := time.Date(2012, time.December, 21, 21, 21, 21, 0, time.UTC)
		n := exqlnull.New(want)
		data, err := json.Marshal(n)
		require.NoError(t, err)
		assert.JSONEq(t, `"2012-12-21T21:21:21Z"`, string(data))

		var decoded exqlnull.Time
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.True(t, decoded.Valid)
		assert.True(t, decoded.V.Equal(want))
	})

	t.Run("null resets value", func(t *testing.T) {
		n := exqlnull.New("before")
		err := json.Unmarshal([]byte(`null`), &n)
		require.NoError(t, err)
		assert.False(t, n.Valid)
		assert.Equal(t, "", n.V)
	})
}

func TestCompatibilityTextUnmarshal(t *testing.T) {
	t.Run("plain string", func(t *testing.T) {
		var n exqlnull.String
		err := n.UnmarshalText([]byte("test"))
		require.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, "test", n.V)
	})

	t.Run("integer", func(t *testing.T) {
		var n exqlnull.Null[int]
		err := n.UnmarshalText([]byte("12345"))
		require.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, 12345, n.V)
	})

	t.Run("float", func(t *testing.T) {
		var n exqlnull.Float64
		err := n.UnmarshalText([]byte("3.14"))
		require.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, 3.14, n.V)
	})

	t.Run("time", func(t *testing.T) {
		var n exqlnull.Time
		err := n.UnmarshalText([]byte("2012-12-21T21:21:21Z"))
		require.NoError(t, err)
		assert.True(t, n.Valid)
		assert.True(t, n.V.Equal(time.Date(2012, time.December, 21, 21, 21, 21, 0, time.UTC)))
	})

	t.Run("empty text becomes null", func(t *testing.T) {
		n := exqlnull.New("before")
		err := n.UnmarshalText([]byte(""))
		require.NoError(t, err)
		assert.False(t, n.Valid)
		assert.Equal(t, "", n.V)
	})

	t.Run("unsupported text returns error", func(t *testing.T) {
		var n exqlnull.Null[struct{ Name string }]
		err := n.UnmarshalText([]byte("anything"))
		require.Error(t, err)
		assert.False(t, n.Valid)
	})
}

func TestCompatibilitySQLRoundTrip(t *testing.T) {
	t.Run("string scan value", func(t *testing.T) {
		var n exqlnull.String
		err := n.Scan("test")
		require.NoError(t, err)
		assert.True(t, n.Valid)
		assert.Equal(t, "test", n.V)

		v, err := n.Value()
		require.NoError(t, err)
		assert.Equal(t, driver.Value("test"), v)
	})

	t.Run("time scan value", func(t *testing.T) {
		want := time.Date(2012, time.December, 21, 21, 21, 21, 0, time.UTC)
		var n exqlnull.Time
		err := n.Scan(want)
		require.NoError(t, err)
		assert.True(t, n.Valid)
		assert.True(t, n.V.Equal(want))

		v, err := n.Value()
		require.NoError(t, err)
		got, ok := v.(time.Time)
		require.True(t, ok)
		assert.True(t, got.Equal(want))
	})

	t.Run("null sql input", func(t *testing.T) {
		var n exqlnull.Null[int]
		err := n.Scan(nil)
		require.NoError(t, err)
		assert.False(t, n.Valid)

		v, err := n.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})
}
