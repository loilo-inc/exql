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

func TestRawPredicate(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		q := RawPredicate("id = ?", 1)
		stmt, args, err := q.Predicate()
		assert.Nil(t, err)
		assert.Equal(t, "id = ?", stmt)
		assert.ElementsMatch(t, []any{1}, args)
	})
	t.Run("should return error if query has no expression", func(t *testing.T) {
		q := RawPredicate("", 1)
		stmt, args, err := q.Predicate()
		assert.EqualError(t, err, "DANGER: empty expression")
		assert.Equal(t, "", stmt)
		assert.Nil(t, args)
	})
}

func TestKeyValuePredicate(t *testing.T) {
	t.Run("should sort columns", func(t *testing.T) {
		now := time.Now()
		clause := KeyValuePredicate(map[string]any{
			"id":         1,
			"created_at": Lt(now),
			"deleted_at": Between("2022-12-03", "2023-01-02"),
			"name":       In("a", "b"),
			"location":   Raw("LIKE ?", "japan"),
		})
		q, args, err := clause.Predicate()
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
		expr := mock_query.NewMockExpression(ctrl)
		expr.EXPECT().Expression(gomock.Any()).Return("", nil, fmt.Errorf("err"))
		clause := KeyValuePredicate(map[string]any{
			"1": expr,
			"2": Eq(1),
		})
		q, args, err := clause.Predicate()
		assert.Equal(t, "", q)
		assert.Nil(t, args)
		assert.ErrorContains(t, err, "err")
	})
	t.Run("should error if one is dangerous query", func(t *testing.T) {
		clause := KeyValuePredicate(map[string]any{
			"id": Raw(""),
		})
		q, args, err := clause.Predicate()
		assert.Equal(t, "", q)
		assert.ErrorContains(t, err, "DANGER")
		assert.Nil(t, args)
	})
}

func TestPredicate_And(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		v := PredicateAnd(
			RawPredicate("`id` = ?", 1),
			RawPredicate("`name` = ?", 2),
			KeyValuePredicate(map[string]any{
				"age": Between(0, 20),
				"cnt": In(3, 4),
			}),
		)
		q, args, err := v.Predicate()
		assert.NoError(t, err)
		assert.Equal(t, "(`id` = ? AND `name` = ? AND (`age` BETWEEN ? AND ? AND `cnt` IN (?,?)))", q)
		assert.ElementsMatch(t, []any{1, 2, 0, 20, 3, 4}, args)
	})
	t.Run("should return error if one returned an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pred := mock_query.NewMockPredicate(ctrl)
		pred.EXPECT().Predicate().Return("", nil, fmt.Errorf("err"))
		and := PredicateAnd(RawPredicate("id = 1"), pred)
		str, args, err := and.Predicate()
		assert.Equal(t, "", str)
		assert.Nil(t, args)
		assert.EqualError(t, err, "err")
	})
}
