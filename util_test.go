package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	var str = "str"
	ptr := Ptr(str)
	assert.Equal(t, "str", *ptr)
}

func TestSyncMap_Load(t *testing.T) {
	t.Run("returns zero value and false when key does not exist", func(t *testing.T) {
		m := &syncMap[string, int]{}

		value, ok := m.Load("missing")

		assert.False(t, ok)
		assert.Zero(t, value)
	})

	t.Run("returns stored value when key exists", func(t *testing.T) {
		m := &syncMap[string, int]{}
		m.Store("answer", 42)

		value, ok := m.Load("answer")

		assert.True(t, ok)
		assert.Equal(t, 42, value)
	})

	t.Run("supports pointer values", func(t *testing.T) {
		type user struct {
			Name string
		}

		m := &syncMap[string, *user]{}
		expected := &user{Name: "alice"}
		m.Store("user", expected)

		value, ok := m.Load("user")

		assert.True(t, ok)
		assert.Same(t, expected, value)
	})
}
