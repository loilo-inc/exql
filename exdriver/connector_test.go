package exdriver_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2/exdriver"
	"github.com/loilo-inc/exql/v2/extest"
	"github.com/loilo-inc/exql/v2/mocks/mock_exdriver"
	"github.com/stretchr/testify/assert"
)

func TestConnector(t *testing.T) {
	conn := exdriver.NewConnector(extest.SqlDB.Driver(), extest.Dsn)
	db := sql.OpenDB(conn)
	assert.NoError(t, db.Ping())
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	hook := mock_exdriver.NewMockQueryHook(ctrl)
	conn.Hooks().Add(hook)
	t.Run("QueryContext", func(t *testing.T) {
		hook.EXPECT().HookQuery(gomock.Any(), "select * from users", []driver.NamedValue{})
		rows, err := db.QueryContext(context.Background(), "select * from users")
		assert.NoError(t, err)
		assert.NotNil(t, rows)
	})
	t.Run("ExecContxet", func(t *testing.T) {
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", []driver.NamedValue{})
		res, err := db.QueryContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
	t.Run("PrepareContext/QueryContext", func(t *testing.T) {
		stmt, err := db.PrepareContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", []driver.NamedValue{})
		rows, err := stmt.QueryContext(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, rows)
	})
	t.Run("PrepareContext/ExecContext", func(t *testing.T) {
		stmt, err := db.PrepareContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", []driver.NamedValue{})
		res, err := stmt.ExecContext(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}
