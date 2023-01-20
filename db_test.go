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
	db.SetDB(extest.SqlDB)
	assert.Equal(t, extest.SqlDB, db.DB())
}

func TestDB_Hooks(t *testing.T) {
	db := extest.DB_Exp
	ctrl := gomock.NewController(t)
	hook := mock_exdriver.NewMockQueryHook(ctrl)
	ctx := context.Background()
	noArgs := []driver.NamedValue{}
	db.Hooks().Add(hook)

	t.Run("QueryContext", func(t *testing.T) {
		hook.EXPECT().HookQuery(ctx, "select * from users", noArgs)
		rows, err := db.DB().QueryContext(ctx, "select * from users")
		assert.NoError(t, err)
		assert.NotNil(t, rows)
	})
	t.Run("ExecContxet", func(t *testing.T) {
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", noArgs)
		res, err := db.DB().QueryContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
	t.Run("PrepareContext/QueryContext", func(t *testing.T) {
		stmt, err := db.DB().PrepareContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", noArgs)
		rows, err := stmt.QueryContext(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, rows)
	})
	t.Run("PrepareContext/ExecContext", func(t *testing.T) {
		stmt, err := db.DB().PrepareContext(ctx, "select count(*) from users")
		assert.NoError(t, err)
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", noArgs)
		res, err := stmt.ExecContext(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
	t.Run("Transaction", func(t *testing.T) {
		hook.EXPECT().HookQuery(ctx, "BEGIN", nil)
		hook.EXPECT().HookQuery(ctx, "select count(*) from users", noArgs)
		hook.EXPECT().HookQuery(ctx, "select * from users", noArgs)
		hook.EXPECT().HookQuery(ctx, "COMMIT", nil)
		err := db.TransactionWithContext(ctx, nil, func(tx exql.Tx) error {
			if _, err := tx.Tx().ExecContext(ctx, "select count(*) from users"); err != nil {
				return err
			}
			_, err := tx.Tx().QueryContext(ctx, "select * from users")
			return err
		})
		assert.NoError(t, err)
	})
}
