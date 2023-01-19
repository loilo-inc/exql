package exdriver_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2/exdriver"
	"github.com/loilo-inc/exql/v2/extest"
	"github.com/loilo-inc/exql/v2/mocks/mock_exdriver"
	"github.com/stretchr/testify/assert"
)

type sliceMatcher struct{ slice []driver.NamedValue }

func (s *sliceMatcher) Matches(x interface{}) bool {
	arr, ok := x.([]driver.NamedValue)
	if !ok {
		return false
	} else if len(arr) != len(s.slice) {
		return false
	}
	for i := 0; i < len(arr); i++ {
		if arr[i].Ordinal != s.slice[i].Ordinal {
			return false
		}
		if !reflect.DeepEqual(arr[i].Value, s.slice[i].Value) {
			return false
		}
	}
	return true
}

func (s *sliceMatcher) String() string {
	return fmt.Sprintf("%v", s.slice)
}

func TestConnector(t *testing.T) {
	conn := exdriver.NewConnector(extest.SqlDB.Driver(), extest.Dsn)
	db := sql.OpenDB(conn)
	assert.NoError(t, db.Ping())
	ctx := context.Background()
	toNamedValues := func(args []any) []driver.NamedValue {
		var res = make([]driver.NamedValue, len(args))
		for i, v := range args {
			res[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
		}
		return res
	}
	argsMatcher := func(args []any) gomock.Matcher {
		return &sliceMatcher{slice: toNamedValues(args)}
	}

	type data struct {
		name  string
		query string
		args  []any
		hook  *mock_exdriver.MockQueryHook
	}
	cases := []data{
		{name: "no args", query: "select * from users"},
		{name: "with args", query: "select * from users where id = ?", args: []any{"id"}},
	}
	bodies := []struct {
		name    string
		subject func(t *testing.T, tt data)
	}{
		{name: "QueryContext", subject: func(t *testing.T, tt data) {
			tt.hook.EXPECT().HookQuery(ctx, tt.query, argsMatcher(tt.args))
			rows, err := db.QueryContext(context.Background(), tt.query, tt.args...)
			assert.NoError(t, err)
			assert.NotNil(t, rows)
		}},
		{name: "ExecContext", subject: func(t *testing.T, tt data) {
			tt.hook.EXPECT().HookQuery(ctx, tt.query, argsMatcher(tt.args))
			res, err := db.QueryContext(ctx, tt.query, tt.args...)
			assert.NoError(t, err)
			assert.NotNil(t, res)
		}},
		{name: "PrepareContext/QueryContext", subject: func(t *testing.T, tt data) {
			stmt, err := db.PrepareContext(ctx, tt.query)
			assert.NoError(t, err)
			tt.hook.EXPECT().HookQuery(ctx, tt.query, argsMatcher(tt.args))
			rows, err := stmt.QueryContext(ctx, tt.args...)
			assert.NoError(t, err)
			assert.NotNil(t, rows)
		}},
		{name: "PrepareContext/ExecContext", subject: func(t *testing.T, tt data) {
			stmt, err := db.PrepareContext(ctx, tt.query)
			assert.NoError(t, err)
			tt.hook.EXPECT().HookQuery(ctx, tt.query, argsMatcher(tt.args))
			res, err := stmt.ExecContext(ctx, tt.args...)
			assert.NoError(t, err)
			assert.NotNil(t, res)
		}},
		{name: "Tx", subject: func(t *testing.T, tt data) {
			tt.hook.EXPECT().HookQuery(ctx, "BEGIN", nil)
			tt.hook.EXPECT().HookQuery(ctx, tt.query, argsMatcher(tt.args))
			tt.hook.EXPECT().HookQuery(ctx, "COMMIT", nil)
			tx, err := db.BeginTx(ctx, nil)
			assert.NoError(t, err)
			_, err = tx.ExecContext(ctx, tt.query, tt.args...)
			assert.NoError(t, err)
			err = tx.Commit()
			assert.NoError(t, err)
		}},
	}
	for _, v := range bodies {
		for _, d := range cases {
			t.Run(v.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				hook := mock_exdriver.NewMockQueryHook(ctrl)
				conn.Hooks().Add(hook)
				t.Cleanup(func() { conn.Hooks().Remove(hook) })
				d.hook = hook
				v.subject(t, d)
			})
		}
	}
}
