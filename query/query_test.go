package query_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2/query"
	q "github.com/loilo-inc/exql/v2/query"
)

func TestQuery(t *testing.T) {
	assertQuery(t, q.Val(1), "?", 1)
	assertQuery(t, q.Vals([]int{1, 2}), "?,?", 1, 2)
	assertQuery(t, q.Cols([]string{"a", "b"}), "`a`,`b`")
	assertQuery(t, q.Q("id = ?", 1), "id = ?", 1)
	assertQuery(t,
		q.Set(map[string]any{"a": 1, "b": 2}),
		"`a` = ?,`b` = ?", 1, 2,
	)
	assertQueryErr(t, q.Q(""), "DANGER: empty query")
	assertQueryErr(t, q.Vals[any](nil), "empty values")
	assertQueryErr(t, q.Cols(nil), "empty columns")
	assertQueryErr(t, q.Set(map[string]any{}), "empty values for set clause")
}

func TestQprintf(t *testing.T) {
	assertQuery(t, query.Qprintf("id in (%s)", query.Vals([]int{1, 2})), "id in (?,?)", 1, 2)
	assertQueryErr(t, query.Qprintf(""), "DANGER: empty query")
	assertQueryErr(t, query.Qprintf("%s", q.Q("")), "DANGER: empty query")
}

func TestCondition(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		cond := q.Cond("id = ?", 1)
		cond.And("name = ?", "go")
		cond.Or("age = ?", 18)
		cond.AndCond(q.Cond("foo = ?", "foo"))
		cond.OrCond(q.Cond("var = ?", "var"))
		assertQuery(t, cond,
			"id = ? AND name = ? OR age = ? AND foo = ? OR var = ?",
			1, "go", 18, "foo", "var",
		)
	})
	t.Run("should error if query retuerned an error", func(t *testing.T) {
		cond := q.CondFrom(q.Q(""))
		assertQueryErr(t, cond, "DANGER: empty query")
	})
}
