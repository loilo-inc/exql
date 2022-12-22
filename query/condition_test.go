package query_test

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

func TestWhere(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		q := NewCondition("id = ?", 1)
		stmt, args, err := q.Condition()
		assert.Nil(t, err)
		assert.Equal(t, "id = ?", stmt)
		assert.ElementsMatch(t, []any{1}, args)
	})
	t.Run("should return error if query has no predicate", func(t *testing.T) {
		q := NewCondition("", 1)
		stmt, args, err := q.Condition()
		assert.EqualError(t, err, "DANGER: empty predicate")
		assert.Equal(t, "", stmt)
		assert.Nil(t, args)
	})
}

func TestWhereEx(t *testing.T) {
	t.Run("should sort columns", func(t *testing.T) {
		now := time.Now()
		clause := NewKeyValueCondition(map[string]any{
			"id":         1,
			"created_at": Lt(now),
			"deleted_at": Between("2022-12-03", "2023-01-02"),
			"name":       In("a", "b"),
			"location":   Raw("LIKE ?", "japan"),
		})
		q, args, err := clause.Condition()
		assert.NoError(t, err)
		preds := []string{
			"`created_at` < ?",
			"`deleted_at` BETWEEN ? AND ?",
			"`id` = ?",
			"`location` LIKE ?",
			"`name` IN (?,?)",
		}

		assert.Equal(t, fmt.Sprintf("(%s)", strings.Join(preds, " AND ")), q)
		assert.ElementsMatch(t, []any{
			1, now, "2022-12-03", "2023-01-02", "a", "b", "japan",
		}, args)
	})
	t.Run("should error if one returned an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		expr := mock_query.NewMockPredicate(ctrl)
		expr.EXPECT().Predicate(gomock.Any()).Return("", nil, fmt.Errorf("err"))
		clause := NewKeyValueCondition(map[string]any{
			"1": expr,
			"2": Eq(1),
		})
		q, args, err := clause.Condition()
		assert.Equal(t, "", q)
		assert.Nil(t, args)
		assert.ErrorContains(t, err, "err")
	})
	t.Run("should error if one is dangerous query", func(t *testing.T) {
		clause := NewKeyValueCondition(map[string]any{
			"id": Raw(""),
		})
		q, args, err := clause.Condition()
		assert.Equal(t, "", q)
		assert.ErrorContains(t, err, "DANGER")
		assert.Nil(t, args)
	})
}

func TestWhereAnd(t *testing.T) {
	t.Run("and", func(t *testing.T) {
		v := ConditionAnd(
			NewCondition("`id` = ?", 1),
			NewCondition("`name` = ?", 2),
			NewKeyValueCondition(map[string]any{
				"age": Between(0, 20),
				"cnt": In(3, 4),
			}),
		)
		q, args, err := v.Condition()
		assert.NoError(t, err)
		assert.Equal(t, "(`id` = ? AND `name` = ? AND (`age` BETWEEN ? AND ? AND `cnt` IN (?,?)))", q)
		assert.ElementsMatch(t, []any{1, 2, 0, 20, 3, 4}, args)
	})
	t.Run("or", func(t *testing.T) {
		v := ConditionOr(
			NewCondition("`id` = ?", 1),
			NewCondition("`name` = ?", 2),
			NewKeyValueCondition(map[string]any{
				"age": Between(0, 20),
				"cnt": In(3, 4),
			}),
		)
		q, args, err := v.Condition()
		assert.NoError(t, err)
		assert.Equal(t, "(`id` = ? OR `name` = ? OR (`age` BETWEEN ? AND ? AND `cnt` IN (?,?)))", q)
		assert.ElementsMatch(t, []any{1, 2, 0, 20, 3, 4}, args)
	})
	t.Run("nest", func(t *testing.T) {
		v, a, err := ConditionAnd(
			ConditionAnd(NewCondition("q = ?", 1), NewCondition("p = ?", 2)),
			ConditionOr(NewCondition("r = ?", 3), NewCondition("s = ?", 4)),
		).Condition()
		assert.NoError(t, err)
		assert.Equal(t, "((q = ? AND p = ?) AND (r = ? OR s = ?))", v)
		assert.ElementsMatch(t, []any{1, 2, 3, 4}, a)

	})
	t.Run("should return error if one returned an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pred := mock_query.NewMockCondition(ctrl)
		pred.EXPECT().Condition().Return("", nil, fmt.Errorf("err"))
		and := ConditionAnd(NewCondition("id = 1"), pred)
		str, args, err := and.Condition()
		assert.Equal(t, "", str)
		assert.Nil(t, args)
		assert.EqualError(t, err, "err")
	})
}
