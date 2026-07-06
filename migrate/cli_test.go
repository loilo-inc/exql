package migrate

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCliFS = fstest.MapFS{
	"1_create_users.up.sql":   {Data: []byte("create table users (id int not null primary key)")},
	"1_create_users.down.sql": {Data: []byte("drop table users")},
}

func TestCli_Run(t *testing.T) {
	ctx := context.Background()
	t.Run("up, version, down and drop", func(t *testing.T) {
		db := testDB(t, "exql_migrate_cli")
		var out bytes.Buffer
		cli := &Cli{DB: db, FS: testCliFS, Out: &out}

		assert.NoError(t, cli.Run(ctx, "up"))
		assert.Equal(t, "applying migration 1_create_users\n", out.String())
		assert.Equal(t, []string{"schema_migrations", "users"}, tableNames(t, db))

		out.Reset()
		assert.NoError(t, cli.Run(ctx, "version"))
		assert.Equal(t, "1\n", out.String())

		assert.NoError(t, cli.Run(ctx, "down"))
		assert.Equal(t, []string{"schema_migrations"}, tableNames(t, db))

		assert.NoError(t, cli.Run(ctx, "drop"))
		assert.Empty(t, tableNames(t, db))
	})
	t.Run("version reports dirty state", func(t *testing.T) {
		db := testDB(t, "exql_migrate_cli_dirty")
		broken := fstest.MapFS{
			"1_broken.up.sql": {Data: []byte("alter table nonexistent add x int")},
		}
		var out bytes.Buffer
		cli := &Cli{DB: db, FS: broken, Out: &out}
		assert.ErrorContains(t, cli.Run(ctx, "up"), "migration 1_broken failed")

		out.Reset()
		assert.NoError(t, cli.Run(ctx, "version"))
		assert.Equal(t, "1 (dirty)\n", out.String())
	})
	t.Run("create", func(t *testing.T) {
		stubTimeNow(t, time.Date(2026, 7, 6, 12, 34, 56, 0, time.UTC))
		dir := t.TempDir()
		var out bytes.Buffer
		// create does not require DB nor FS
		cli := &Cli{Dir: dir, Out: &out}
		assert.NoError(t, cli.Run(ctx, "create", "add_age"))
		assert.Equal(t,
			filepath.Join(dir, "20260706123456_add_age.up.sql")+"\n"+
				filepath.Join(dir, "20260706123456_add_age.down.sql")+"\n",
			out.String())

		assert.EqualError(t, cli.Run(ctx, "create"), "usage: create <name>")
		assert.EqualError(t, cli.Run(ctx, "create", "a", "b"), "usage: create <name>")
	})
	t.Run("create defaults to the migrations directory", func(t *testing.T) {
		stubTimeNow(t, time.Date(2026, 7, 6, 12, 34, 56, 0, time.UTC))
		dir := t.TempDir()
		t.Chdir(dir)
		require.NoError(t, (&Cli{Out: &bytes.Buffer{}}).Run(ctx, "create", "add_age"))
		assert.FileExists(t, filepath.Join(dir, "migrations", "20260706123456_add_age.up.sql"))
	})
	t.Run("errors", func(t *testing.T) {
		cli := &Cli{FS: testCliFS, Out: &bytes.Buffer{}}
		assert.ErrorContains(t, cli.Run(ctx), "expects a subcommand")
		assert.ErrorContains(t, cli.Run(ctx, "unknown"), "unknown command: unknown")

		broken := &Cli{FS: fstest.MapFS{"1_a.down.sql": {Data: []byte("x")}}, Out: &bytes.Buffer{}}
		assert.EqualError(t, broken.Run(ctx, "up"), "missing up migration for 1_a")
	})
}
