package exql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestWhereEx(t *testing.T) {
	WhereEx(map[string]any{
		"id":         1,
		"created_at": Lt(time.Now()),
		"deleted_at": Between("2022-12-03", "2023-01-02"),
		"age":        Range(Gte(0), Lt(20)),
		"name":       In("a", "b"),
		"location":   Raw("= ?", "japan"),
	})
}
