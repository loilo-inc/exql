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

func TestConditions_Add(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		c := NewConditions([]*Condition{})
		assert.False(t, c.IsSafe())
		c.Add(&Condition{Text: "id = ?", Arg: 1})
		assert.True(t, c.IsSafe())
	})
}

func TestConditions_String(t *testing.T) {
	t.Run("should return error when not safe conditions", func(t *testing.T) {
		c := NewConditions([]*Condition{})
		_, err := c.String(nil)
		assert.EqualError(t, err, "no conditions")
	})
	t.Run("should return string concatenated with and", func(t *testing.T) {
		c := NewConditions([]*Condition{
			{Text: "id = ?", Arg: 1},
			{Text: "flag = true"},
		})
		result, err := c.String(nil)
		assert.Nil(t, err)
		assert.Equal(t, "id = ? and flag = true", result)
	})
	t.Run("add prefix", func(t *testing.T) {
		c := NewConditions([]*Condition{
			{Text: "id = ?", Arg: 1},
			{Text: "flag = true"},
		})
		prefix := "users"
		result, err := c.String(&prefix)
		assert.Nil(t, err)
		assert.Equal(t, "users.id = ? and users.flag = true", result)
	})
}

func TestConditions_Args(t *testing.T) {
	t.Run("should return slice of args", func(t *testing.T) {
		c := NewConditions([]*Condition{
			{Text: "id = ?", Arg: 1},
			{Text: "flag = true"},
			{Text: "number = ?", Arg: 2},
		})
		args := c.Args()
		assert.Equal(t, []interface{}{1, 2}, args)
	})
}
