package query_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2/query"
)

func TestBuilder(t *testing.T) {
	assertQuery := func(q query.Query, str string, args ...any) {
	}
	assertQuery(
		query.NewBuilder().Sprintf("this is %s", "str").Build(),
		"this is str",
	)
	assertQuery(
		query.NewBuilder().Qprintf("(%s)", query.NewQuery("id = ?", 1)).Build(),
		"(id = ?)", 1,
	)
	assertQuery(
		query.NewBuilder().Query("id = ?", 1).Build(),
		"id = ?", 1,
	)
	assertQuery(
		query.NewBuilder().Add(query.NewQuery("id = ?", 1)).Build(),
		"id = ?", 1,
	)
	assertQuery(
		query.NewBuilder().Query("?,?").Args(1, 2).Build(),
		"?,?", 1, 2,
	)
}
