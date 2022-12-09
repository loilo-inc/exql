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

func TestNewStmt(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		q := NewStmt("id = ?", 1)
		stmt, args, err := q.Stmt()
		assert.Nil(t, err)
		assert.Equal(t, "id = ?", stmt)
		assert.ElementsMatch(t, []any{1}, args)
	})
	t.Run("should return error if query has no expression", func(t *testing.T) {
		q := NewStmt("", 1)
		stmt, args, err := q.Stmt()
		assert.EqualError(t, err, "DANGER: empty where clause")
		assert.Equal(t, "", stmt)
		assert.Nil(t, args)
	})
}

func TestNewStmtEx(t *testing.T) {
	t.Run("should sort columns", func(t *testing.T) {
		now := time.Now()
		clause := NewStmtEx(map[string]any{
			"id":         1,
			"created_at": Lt(now),
			"deleted_at": Between("2022-12-03", "2023-01-02"),
			"name":       In("a", "b"),
			"location":   Raw("LIKE ?", "japan"),
		})
		q, args, err := clause.Stmt()
		assert.NoError(t, err)
		stmt := []string{
			"`created_at` < ?",
			"`deleted_at` BETWEEN ? AND ?",
			"`id` = ?",
			"`location` LIKE ?",
			"`name` IN (?,?)",
		}
		assert.Equal(t, strings.Join(stmt, " AND "), q)
		assert.ElementsMatch(t, []any{
			1, now, "2022-12-03", "2023-01-02", "a", "b", "japan",
		}, args)
	})
	t.Run("should error if one returned an error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		expr := mock_query.NewMockExpr(ctrl)
		expr.EXPECT().Expr(gomock.Any()).Return("", nil, fmt.Errorf("err"))
		clause := NewStmtEx(map[string]any{
			"1": expr,
			"2": Eq(1),
		})
		q, args, err := clause.Stmt()
		assert.Equal(t, "", q)
		assert.Nil(t, args)
		assert.ErrorContains(t, err, "err")
	})
	t.Run("should error if one is dangerous query", func(t *testing.T) {
		clause := NewStmtEx(map[string]any{
			"id": Raw(""),
		})
		q, args, err := clause.Stmt()
		assert.Equal(t, "", q)
		assert.Equal(t, err, ErrDangerousExpr)
		assert.Nil(t, args)
	})
}

func TestStmtAnd(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		v := ConcatStmt(
			NewStmt("`id` = ?", 1),
			NewStmt("`name` = ?", 2),
			NewStmtEx(map[string]any{
				"age": Between(0, 20),
				"cnt": In(3, 4),
			}),
		)
		q, args, err := v.Stmt()
		assert.NoError(t, err)
		assert.Equal(t, "(`id` = ?) AND (`name` = ?) AND (`age` BETWEEN ? AND ? AND `cnt` IN (?,?))", q)
		assert.ElementsMatch(t, []any{1, 2, 0, 20, 3, 4}, args)
	})
	t.Run("should error if one returned an error", func(t *testing.T) {
		stmt, args, err := ConcatStmt(NewStmt("id = ?", 1), NewStmt("")).Stmt()
		assert.Nil(t, args)
		assert.Equal(t, "", stmt)
		assert.ErrorContains(t, err, "DANGER")
	})
}
