package exql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsSafeWhereClause(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.False(t, IsSafeWhereClause(""))
	})
	t.Run("space", func(t *testing.T) {
		assert.False(t, IsSafeWhereClause("  "))
	})
	t.Run("tag", func(t *testing.T) {
		assert.False(t, IsSafeWhereClause("\t\t"))
	})
	t.Run("space and tag", func(t *testing.T) {
		assert.False(t, IsSafeWhereClause(" \t"))
	})
}

func TestWhereQuery_Query(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		q := Where("id = ?", 1)
		act, err := q.Query()
		assert.Nil(t, err)
		assert.Equal(t, "id = ?", act)
	})
	t.Run("should return error if query has no expression", func(t *testing.T) {
		q := Where("", 1)
		_, err := q.Query()
		assert.EqualError(t, err, "DANGER: empty where clause")
	})
}

func TestWhereQuery_Args(t *testing.T) {
	w := Where("id = ?", 1, 2)
	args := w.Args()
	assert.ElementsMatch(t, []interface{}{1, 2}, args)
}
