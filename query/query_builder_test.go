package query

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/mocks/mock_query"
	"github.com/stretchr/testify/assert"
)

func TestSqlPraceholder(t *testing.T) {
	assert.Equal(t, "", SqlPlaceHolders(0))
	assert.Equal(t, "?", SqlPlaceHolders(1))
	assert.Equal(t, "?,?,?", SqlPlaceHolders(3))
}

func TestQuery(t *testing.T) {
	tt := func(expr Expr, query string, args ...any) {
		v, err := expr.Expr("a")
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("`a` %s", query), v)
		assert.ElementsMatch(t, args, expr.Args())
	}
	tt(Eq(1), "= ?", 1)
	tt(NotEq(1), "!= ?", 1)
	tt(IsNull(), "IS NULL")
	tt(IsNotNull(), "IS NOT NULL")
	tt(Lt(0), "< ?", 0)
	tt(Lte(0), "<= ?", 0)
	tt(Gt(0), "> ?", 0)
	tt(Gte(0), ">= ?", 0)
	tt(In(0, 1), "IN (?,?)", 0, 1)
	tt(In([]int{0, 1}...), "IN (?,?)", 0, 1)
}

func TestQuery_And(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Run("basic", func(t *testing.T) {
		expr := And(Eq(1), NotEq(2))
		v, err := expr.Expr("a")
		assert.NoError(t, err)
		assert.Equal(t, "`a` = ? AND `a` != ?", v)
		assert.ElementsMatch(t, []any{1, 2}, expr.Args())
	})
	t.Run("should error if one returned error", func(t *testing.T) {
		expr := mock_query.NewMockExpr(ctrl)
		expr.EXPECT().Expr("a").Return("", fmt.Errorf("err"))
		v, err := And(expr, Eq(1)).Expr("a")
		assert.Equal(t, "", v)
		assert.ErrorContains(t, err, "err")
	})
}

func TestQuery_Between(t *testing.T) {
	expr := Between(1, 2)
	v, err := expr.Expr("a")
	assert.NoError(t, err)
	assert.Equal(t, "`a` BETWEEN ? AND ?", v)
	assert.ElementsMatch(t, []any{1, 2}, expr.Args())
}

func TestIsSafeWhereClause(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		for _, v := range []string{"", " ", "\t", " \t", "\n"} {
			_, err := GuardDangerousQuery(v)
			assert.Equal(t, ErrDangerousExpr, err)
		}
	})
	t.Run("safe", func(t *testing.T) {
		q, err := GuardDangerousQuery("a")
		assert.Equal(t, "a", q)
		assert.NoError(t, err)
	})
}
