package query_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
)

func assertQuery(t *testing.T, q query.Query, str string, args ...any) {
	stmt, vals, err := q.Query()
	assert.NoError(t, err)
	assert.Equal(t, str, stmt)
	assert.ElementsMatch(t, args, vals)
}
func assertQueryErr(t *testing.T, q query.Query, msg string) {
	stmt, vals, err := q.Query()
	assert.EqualError(t, err, msg)
	assert.Equal(t, "", stmt)
	assert.Nil(t, vals)
}
func TestBuilder(t *testing.T) {
	t.Run("Sprintf", func(t *testing.T) {
		assertQuery(t,
			query.NewBuilder().Sprintf("this is %s", "str").Build(),
			"this is str",
		)
	})
	t.Run("Qprintf", func(t *testing.T) {
		assertQuery(t,
			query.NewBuilder().Qprintf("(%s)", query.Q("id = ?", 1)).Build(),
			"(id = ?)", 1,
		)
	})
	t.Run("Query", func(t *testing.T) {
		assertQuery(t,
			query.NewBuilder().Query("id = ?", 1).Build(),
			"id = ?", 1,
		)
	})
	t.Run("Add", func(t *testing.T) {
		assertQuery(t,
			query.NewBuilder().Add(query.Q("id = ?", 1)).Build(),
			"id = ?", 1,
		)
	})
	t.Run("Args", func(t *testing.T) {
		assertQuery(t,
			query.NewBuilder().Query("?,?").Args(1, 2).Build(),
			"?,?", 1, 2,
		)
	})
	t.Run("Clone", func(t *testing.T) {
		base := query.NewBuilder().Query("id = ?", 1)
		copy := base.Clone()
		assertQuery(t, base.Build(), "id = ?", 1)
		assertQuery(t, copy.Build(), "id = ?", 1)
	})
}
