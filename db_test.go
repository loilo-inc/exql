package exql_test

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/extest"
	"github.com/loilo-inc/exql/v2/mocks/mock_exdriver"
	"github.com/stretchr/testify/assert"
)

func TestNewDB(t *testing.T) {
	d := extest.SqlDB
	db := exql.NewDB(d)
	assert.Equal(t, d, db.DB())
	t.Run("hooks shoub be disabled", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"hooks is disabled because there's no hooked connector",
			func() { db.Hooks() },
		)
	})
}

func TestDB_SetDB(t *testing.T) {
	db := extest.TestDb()
	assert.NotNil(t, db.Hooks())
	db.SetDB(extest.SqlDB)
	assert.PanicsWithValue(t,
		"hooks is disabled because there's no hooked connector",
		func() { db.Hooks() },
	)
}

func TestDB_Hooks(t *testing.T) {
	db := extest.TestDb()
	ctrl := gomock.NewController(t)
	hook := mock_exdriver.NewMockQueryHook(ctrl)
	ctx := context.Background()
	db.Hooks().Add(hook)

	t.Run("QueryContext", func(t *testing.T) {
		hook.EXPECT().HookQuery(ctx, "select * from users", []driver.NamedValue{})
		rows, err := db.DB().QueryContext(ctx, "select * from users")
		assert.NoError(t, err)
		assert.NotNil(t, rows)
	})
	t.Run("ExecContxet", func(t *testing.T) {
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", []driver.NamedValue{})
		res, err := db.DB().QueryContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
	t.Run("PrepareContext/QueryContext", func(t *testing.T) {
		stmt, err := db.DB().PrepareContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", []driver.NamedValue{})
		rows, err := stmt.QueryContext(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, rows)
	})
	t.Run("PrepareContext/ExecContext", func(t *testing.T) {
		stmt, err := db.DB().PrepareContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", []driver.NamedValue{})
		res, err := stmt.ExecContext(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}
