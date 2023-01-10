package query_test

import (
	"testing"

	q "github.com/loilo-inc/exql/v2/query"
)

func TestQuery(t *testing.T) {
	assertQuery(t, q.V(1, 2), "?,?", 1, 2)
	assertQuery(t, q.Vals([]int{1, 2}), "?,?", 1, 2)
	assertQuery(t, q.Cols("a.b", "c.*"), "`a`.`b`,`c`.*")
	assertQuery(t, q.Q("id = ?", 1), "id = ?", 1)
	assertQuery(t,
		q.Set(map[string]any{"a": 1, "b": 2}),
		"`a` = ?,`b` = ?", 1, 2,
	)
	assertQueryErr(t, q.Q(""), "DANGER: empty query")
	assertQueryErr(t, q.Vals[any](nil), "empty values")
	assertQueryErr(t, q.Cols(), "empty columns")
	assertQueryErr(t, q.Set(map[string]any{}), "empty values for set clause")
}

func TestNew(t *testing.T) {
	assertQuery(t,
		q.New("id in (:?) and name = ? and more", q.Vals([]int{1, 2}), "go"),
		"id in (?,?) and name = ? and more", 1, 2, "go",
	)
	assertQueryErr(t, q.New(""), "DANGER: empty query")
	assertQueryErr(t, q.New(":?", q.Q("")), "DANGER: empty query")
	assertQueryErr(t, q.New("?"), "missing argument at 0")
	assertQueryErr(t, q.New("?,?", 1), "missing argument at 1")
	assertQueryErr(t, q.New(":?", 1), "unexpected argument type for :? placeholder at 0")
	assertQueryErr(t, q.New("?", 1, 2), "arguments count mismatch: found 1, got 2")
}

func TestCondition(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cond := q.Cond("id = ?", 1)
		cond.And("name = ?", "go")
		cond.Or("age in (:?)", q.V(20, 21))
		cond.AndCond(q.Cond("foo = ?", "foo"))
		cond.OrCond(q.Cond("var = ?", "var"))
		assertQuery(t, cond,
			"id = ? AND name = ? OR age in (?,?) AND foo = ? OR var = ?",
			1, "go", 20, 21, "foo", "var",
		)
	})
	t.Run("should error if query retuerned an error", func(t *testing.T) {
		cond := q.CondFrom(q.Q(""))
		assertQueryErr(t, cond, "DANGER: empty query")
	})
}
