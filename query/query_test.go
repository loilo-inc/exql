package query_test

import (
	"testing"

	q "github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
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

func TestCondition(t *testing.T) {
	cond := q.Cond("id = ?", 1)
	cond.And("name = ?", "go")
	cond.Or("age = ?", 18)
	cond.AndCond(q.Cond("foo = ?", "foo"))
	cond.OrCond(q.Cond("var = ?", "var"))
	stmt, args, err := cond.Query()
	assert.NoError(t, err)
	assert.Equal(t, "id = ? AND name = ? OR age = ? AND foo = ? OR var = ?", stmt)
	assert.ElementsMatch(t, []any{1, "go", 18, "foo", "var"}, args)
}
