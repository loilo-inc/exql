package query_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2/mocks/mock_query"
	. "github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
)

func TestPredicate(t *testing.T) {
	tt := func(pred Predicate, query string, args ...any) {
		v, a, err := pred.Predicate("a")
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
	tt(Raw("SOME ?", 1), "SOME ?", 1)
	tt(In([]int{0, 1}...), "IN (?,?)", 0, 1)
}

func TestPredicateAnd(t *testing.T) {
	and := PredicateAnd(Eq(1), NotEq(2))
	or := PredicateOr(Eq(3), NotEq(4))
	t.Run("and", func(t *testing.T) {
		v, a, err := and.Predicate("a")
		assert.NoError(t, err)
		assert.Equal(t, "(`a` = ? AND `a` <> ?)", v)
		assert.ElementsMatch(t, []any{1, 2}, a)
	})
	t.Run("or", func(t *testing.T) {
		v, a, err := or.Predicate("a")
		assert.NoError(t, err)
		assert.Equal(t, "(`a` = ? OR `a` <> ?)", v)
		assert.ElementsMatch(t, []any{3, 4}, a)
	})
	t.Run("nest", func(t *testing.T) {
		v, a, err := PredicateAnd(and, or).Predicate("a")
		assert.NoError(t, err)
		assert.Equal(t, "((`a` = ? AND `a` <> ?) AND (`a` = ? OR `a` <> ?))", v)
		assert.ElementsMatch(t, []any{1, 2, 3, 4}, a)
	})
	t.Run("should error if one returned error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		expr := mock_query.NewMockPredicate(ctrl)
		expr.EXPECT().Predicate("a").Return("", nil, fmt.Errorf("err"))
		v, a, err := PredicateAnd(expr, Eq(1)).Predicate("a")
		assert.Equal(t, "", v)
		assert.Nil(t, a)
		assert.ErrorContains(t, err, "err")
	})
}

func TestExpr_Between(t *testing.T) {
	pred := Between(1, 2)
	v, a, err := pred.Predicate("a")
	assert.NoError(t, err)
	assert.Equal(t, "`a` BETWEEN ? AND ?", v)
	assert.ElementsMatch(t, []any{1, 2}, a)
}
