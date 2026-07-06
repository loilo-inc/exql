package migrate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testMigrations = []*Migration{
	{
		Version:  1,
		Name:     "create_users",
		UpStmt:   "create table users (id int not null primary key, name varchar(255) not null)",
		DownStmt: "drop table users",
		HasDown:  true,
	},
	{
		Version:  2,
		Name:     "add_age",
		UpStmt:   "alter table users add age int not null default 0",
		DownStmt: "alter table users drop age",
		HasDown:  true,
	},
}

func TestMigrator_Up(t *testing.T) {
	ctx := context.Background()
	t.Run("applies all migrations and is idempotent", func(t *testing.T) {
		db := testDB(t, "exql_migrate_up")
		var logs []string
		m := New(db, testMigrations, &Options{
			Log: func(msg string) { logs = append(logs, msg) },
		})

		assert.NoError(t, m.Up(ctx))
		assert.Equal(t, []string{"schema_migrations", "users"}, tableNames(t, db))
		assert.Equal(t, []string{
			"applying migration 1_create_users",
			"applying migration 2_add_age",
		}, logs)

		version, dirty, err := m.Version(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), version)
		assert.False(t, dirty)

		// second run applies nothing
		logs = nil
		assert.NoError(t, m.Up(ctx))
		assert.Empty(t, logs)
	})
	t.Run("applies only migrations newer than the current version", func(t *testing.T) {
		db := testDB(t, "exql_migrate_up_partial")
		m := New(db, testMigrations[:1], nil)
		require.NoError(t, m.Up(ctx))

		var logs []string
		m = New(db, testMigrations, &Options{
			Log: func(msg string) { logs = append(logs, msg) },
		})
		assert.NoError(t, m.Up(ctx))
		assert.Equal(t, []string{"applying migration 2_add_age"}, logs)
	})
	t.Run("applies blank migrations as no-ops", func(t *testing.T) {
		db := testDB(t, "exql_migrate_up_blank")
		blank := []*Migration{{Version: 1, Name: "blank", UpStmt: "\n", HasDown: true}}
		m := New(db, blank, nil)
		assert.NoError(t, m.Up(ctx))

		version, dirty, err := m.Version(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), version)
		assert.False(t, dirty)

		assert.NoError(t, m.Down(ctx))
	})
	t.Run("marks the database dirty on failure", func(t *testing.T) {
		db := testDB(t, "exql_migrate_up_dirty")
		broken := append(testMigrations[:1:1], &Migration{
			Version: 2, Name: "broken", UpStmt: "alter table nonexistent add x int",
		})
		m := New(db, broken, nil)
		err := m.Up(ctx)
		assert.ErrorContains(t, err, "migration 2_broken failed")

		version, dirty, err := m.Version(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), version)
		assert.True(t, dirty)

		// further migrations are rejected until fixed manually
		assert.ErrorContains(t, m.Up(ctx), "database is dirty at version 2")
		assert.ErrorContains(t, m.Down(ctx), "database is dirty at version 2")
	})
}

func TestMigrator_Down(t *testing.T) {
	ctx := context.Background()
	t.Run("reverts all migrations in reverse order", func(t *testing.T) {
		db := testDB(t, "exql_migrate_down")
		var logs []string
		m := New(db, testMigrations, &Options{
			Log: func(msg string) { logs = append(logs, msg) },
		})
		require.NoError(t, m.Up(ctx))

		logs = nil
		assert.NoError(t, m.Down(ctx))
		assert.Equal(t, []string{
			"reverting migration 2_add_age",
			"reverting migration 1_create_users",
		}, logs)
		assert.Equal(t, []string{"schema_migrations"}, tableNames(t, db))

		version, dirty, err := m.Version(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(-1), version)
		assert.False(t, dirty)

		// up applies everything again after down
		assert.NoError(t, m.Up(ctx))
		assert.Equal(t, []string{"schema_migrations", "users"}, tableNames(t, db))
	})
	t.Run("fails if a down migration is missing", func(t *testing.T) {
		db := testDB(t, "exql_migrate_down_missing")
		noDown := []*Migration{{
			Version: 1, Name: "create_users", UpStmt: testMigrations[0].UpStmt,
		}}
		m := New(db, noDown, nil)
		require.NoError(t, m.Up(ctx))
		assert.EqualError(t, m.Down(ctx), "missing down migration for 1_create_users")
	})
}

func TestMigrator_Drop(t *testing.T) {
	ctx := context.Background()
	t.Run("drops all tables", func(t *testing.T) {
		db := testDB(t, "exql_migrate_drop")
		m := New(db, testMigrations, nil)
		require.NoError(t, m.Up(ctx))

		assert.NoError(t, m.Drop(ctx))
		assert.Empty(t, tableNames(t, db))
	})
	t.Run("does nothing on an empty database", func(t *testing.T) {
		db := testDB(t, "exql_migrate_drop_empty")
		m := New(db, nil, nil)
		assert.NoError(t, m.Drop(ctx))
	})
	t.Run("drops tables referenced by foreign keys", func(t *testing.T) {
		db := testDB(t, "exql_migrate_drop_fk")
		fk := []*Migration{
			{Version: 1, Name: "parent", UpStmt: "create table parents (id int not null primary key)"},
			{Version: 2, Name: "child", UpStmt: "create table children (id int not null primary key, parent_id int not null, foreign key (parent_id) references parents (id))"},
		}
		m := New(db, fk, nil)
		require.NoError(t, m.Up(ctx))

		assert.NoError(t, m.Drop(ctx))
		assert.Empty(t, tableNames(t, db))
	})
}

func TestMigrator_Version(t *testing.T) {
	ctx := context.Background()
	t.Run("returns -1 before any migration", func(t *testing.T) {
		db := testDB(t, "exql_migrate_version")
		m := New(db, testMigrations, nil)
		version, dirty, err := m.Version(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(-1), version)
		assert.False(t, dirty)
	})
}

func TestMigrator_Lock(t *testing.T) {
	ctx := context.Background()
	t.Run("fails if the lock is held by another session", func(t *testing.T) {
		db := testDB(t, "exql_migrate_lock")
		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()
		var acquired int
		require.NoError(t, conn.QueryRowContext(ctx, "SELECT GET_LOCK('exql-migrate:exql_migrate_lock', 0)").Scan(&acquired))
		require.Equal(t, 1, acquired)
		defer conn.ExecContext(ctx, "SELECT RELEASE_LOCK('exql-migrate:exql_migrate_lock')")

		m := New(db, testMigrations, &Options{LockTimeout: time.Second})
		assert.ErrorContains(t, m.Up(ctx), `failed to acquire lock "exql-migrate:exql_migrate_lock"`)
	})
	t.Run("uses a custom lock name", func(t *testing.T) {
		db := testDB(t, "exql_migrate_lock_custom")
		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()
		var acquired int
		require.NoError(t, conn.QueryRowContext(ctx, "SELECT GET_LOCK('custom-lock', 0)").Scan(&acquired))
		require.Equal(t, 1, acquired)
		defer conn.ExecContext(ctx, "SELECT RELEASE_LOCK('custom-lock')")

		m := New(db, testMigrations, &Options{LockName: "custom-lock", LockTimeout: time.Second})
		assert.ErrorContains(t, m.Up(ctx), `failed to acquire lock "custom-lock"`)
	})
}
