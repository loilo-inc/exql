package query_test

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
)

type desc struct {
	query Query
	stmt  string
	args  []any
}

type desce struct {
	query Query
	err   string
}

func TestQueryBuilder(t *testing.T) {
	arr := []desc{
		{
			query: Insert{
				Into: "table",
				Values: map[string]any{
					"id": 1,
				},
			},
			stmt: "INSERT INTO `table` (`id`) VALUES (?)",
			args: []any{1},
		},
		{
			query: InsertMany{
				Into:    "table",
				Columns: []string{"id"},
				Values: [][]any{
					{1}, {2}, {3},
				},
			},
			stmt: "INSERT INTO `table` (`id`) VALUES (?),(?),(?)",
			args: []any{1, 2, 3},
		},
		{
			query: Select{
				From:  "table",
				Where: NewCondition("id = ?", 1),
			},
			stmt: "SELECT * FROM `table` WHERE id = ?",
			args: []any{1},
		},
		{
			query: Select{
				From:      "table",
				Columns:   []string{"id", "age"},
				Where:     NewCondition("id = ?", 1),
				Limit:     2,
				Offset:    3,
				OrderBy:   "id DESC",
				ForUpdate: true,
			},
			stmt: "SELECT `id`,`age` FROM `table` WHERE id = ? ORDER BY id DESC LIMIT ? OFFSET ? FOR UPDATE",
			args: []any{1, 2, 3},
		},
		{
			query: Update{
				Table: "table",
				Set: map[string]any{
					"id": 1,
				},
				Where: NewCondition(`id = ?`, 2),
			},
			stmt: "UPDATE `table` SET `id` = ? WHERE id = ?",
			args: []any{1, 2},
		},
		{
			query: Update{
				Table: "table",
				Set: map[string]any{
					"id":   1,
					"name": "go",
				},
				Where:   NewCondition(`id = ?`, 2),
				OrderBy: "id",
				Limit:   3,
				Offset:  4,
			},
			stmt: "UPDATE `table` SET `id` = ?,`name` = ? WHERE id = ? ORDER BY id LIMIT ? OFFSET ?",
			args: []any{1, "go", 2, 3, 4},
		},
		{
			query: Delete{
				From:    "table",
				Where:   NewCondition(`id = ?`, 1),
				OrderBy: "id",
				Limit:   2,
				Offset:  3,
			},
			stmt: "DELETE FROM `table` WHERE id = ? ORDER BY id LIMIT ? OFFSET ?",
			args: []any{1, 2, 3},
		},
	}
	for _, v := range arr {
		t.Run(v.stmt, func(t *testing.T) {
			stmt, args, err := v.query.Query()
			assert.NoError(t, err)
			assert.Equal(t, v.stmt, stmt)
			assert.ElementsMatch(t, v.args, args)
		})
	}
}

func TestBuilderError(t *testing.T) {
	arr := []desce{
		{
			query: Insert{},
			err:   "empty table",
		},
		{
			query: Insert{Into: "table"},
			err:   "empty values",
		},
		{
			query: InsertMany{},
			err:   "empty table",
		},
		{
			query: InsertMany{Into: "table"},
			err:   "empty values",
		},
		{
			query: InsertMany{Into: "table", Columns: []string{"id"}},
			err:   "empty values",
		},
		{
			query: InsertMany{Into: "table", Columns: []string{"id"}, Values: [][]any{{1}, {1, 2}}},
			err:   "number of columns/values mismatch",
		},
		{
			query: Select{},
			err:   "empty table",
		},
		{
			query: Select{From: "table"},
			err:   "empty where clause",
		},
		{
			query: Select{From: "table", Where: NewCondition("")},
			err:   "DANGER",
		},
		{
			query: Update{},
			err:   "empty table",
		},
		{
			query: Update{Table: "table"},
			err:   "empty values",
		},
		{
			query: Update{Table: "table", Set: map[string]any{"id": 1}},
			err:   "empty where clause",
		},
		{
			query: Update{Table: "table", Set: map[string]any{"id": 1}, Where: NewCondition("")},
			err:   "DANGER",
		},
		{
			query: Delete{},
			err:   "empty table",
		},
		{
			query: Delete{From: "table"},
			err:   "empty where clause",
		},
		{
			query: Delete{From: "table", Where: NewCondition("")},
			err:   "DANGER",
		},
	}
	for _, v := range arr {
		s := reflect.TypeOf(v.query).Name()
		t.Run(fmt.Sprintf("%s:%s", s, v.err), func(t *testing.T) {
			stmt, args, err := v.query.Query()
			assert.Equal(t, "", stmt)
			assert.Nil(t, args)
			assert.ErrorContains(t, err, v.err)
		})
	}
}