package exql

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/loilo-inc/exql/mocks/mock_query"
	. "github.com/loilo-inc/exql/query"
)

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
	t.Run("should sort columns", func(t *testing.T) {
		now := time.Now()
		clause := WhereEx(map[string]any{
			"id":         1,
			"created_at": Lt(now),
			"deleted_at": Between("2022-12-03", "2023-01-02"),
			"name":       In("a", "b"),
			"location":   Raw("LIKE ?", "japan"),
		})
		q, err := clause.Query()
		assert.NoError(t, err)
		stmt := []string{
			"(`created_at` < ?)",
			"(`deleted_at` BETWEEN ? AND ?)",
			"(`id` = ?)",
			"(`location` LIKE ?)",
			"(`name` IN (?,?))",
		}
		assert.Equal(t, strings.Join(stmt, " AND "), q)
		assert.ElementsMatch(t, []any{
			1, now, "2022-12-03", "2023-01-02", "a", "b", "japan",
		}, clause.Args())
	})
	t.Run("should error if one returned an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		expr := mock_query.NewMockExpr(ctrl)
		expr.EXPECT().Expr(gomock.Any()).Return("", fmt.Errorf("err"))
		clause := WhereEx(map[string]any{
			"1": expr,
			"2": Eq(1),
		})
		q, err := clause.Query()
		assert.Equal(t, "", q)
		assert.ErrorContains(t, err, "err")
	})
	t.Run("should error if one is dangerous query", func(t *testing.T) {
		clause := WhereEx(map[string]any{
			"id": Raw(""),
		})
		q, err := clause.Query()
		assert.Equal(t, "", q)
		assert.Equal(t, err, ErrDangerousExpr)
	})
}
