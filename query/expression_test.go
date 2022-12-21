package query_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/mocks/mock_query"
	. "github.com/loilo-inc/exql/query"
	"github.com/stretchr/testify/assert"
)

func TestExpr(t *testing.T) {
	tt := func(expr Expression, query string, args ...any) {
		v, a, err := expr.Expression("a")
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("`a` %s", query), v)
		assert.ElementsMatch(t, args, a)
	}
	tt(Eq(1), "= ?", 1)
	tt(NotEq(1), "<> ?", 1)
	tt(IsNull(), "IS NULL")
	tt(IsNotNull(), "IS NOT NULL")
	tt(Like("go"), "LIKE ?", "go")
	tt(Lt(0), "< ?", 0)
	tt(Lte(0), "<= ?", 0)
	tt(Gt(0), "> ?", 0)
	tt(Gte(0), ">= ?", 0)
	tt(In(0, 1), "IN (?,?)", 0, 1)
	tt(In([]int{0, 1}...), "IN (?,?)", 0, 1)
}

func TestExpr_And(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Run("basic", func(t *testing.T) {
		expr := ExpressionAnd(Eq(1), NotEq(2))
		v, a, err := expr.Expression("a")
		assert.NoError(t, err)
		assert.Equal(t, "(`a` = ? AND `a` <> ?)", v)
		assert.ElementsMatch(t, []any{1, 2}, a)
	})
	t.Run("should error if one returned error", func(t *testing.T) {
		expr := mock_query.NewMockExpression(ctrl)
		expr.EXPECT().Expression("a").Return("", nil, fmt.Errorf("err"))
		v, a, err := ExpressionAnd(expr, Eq(1)).Expression("a")
		assert.Equal(t, "", v)
		assert.Nil(t, a)
		assert.ErrorContains(t, err, "err")
	})
}

func TestExpr_Between(t *testing.T) {
	expr := Between(1, 2)
	v, a, err := expr.Expression("a")
	assert.NoError(t, err)
	assert.Equal(t, "`a` BETWEEN ? AND ?", v)
	assert.ElementsMatch(t, []any{1, 2}, a)
}
