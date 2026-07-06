package main

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	t.Run("migrates a database from files in -dir", func(t *testing.T) {
		root, err := sql.Open("mysql", "root:@tcp(127.0.0.1:13326)/")
		require.NoError(t, err)
		defer root.Close()
		_, err = root.Exec("CREATE DATABASE IF NOT EXISTS exql_migrate_main")
		require.NoError(t, err)
		defer func() {
			_, err := root.Exec("DROP DATABASE IF EXISTS exql_migrate_main")
			require.NoError(t, err)
		}()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "1_create_users.up.sql"),
			[]byte("create table users (id int not null primary key)"), 0644))
		dsn := "root:@tcp(127.0.0.1:13326)/exql_migrate_main"

		var out bytes.Buffer
		assert.NoError(t, run(ctx, []string{"-dsn", dsn, "-dir", dir, "up"}, &out))
		assert.Equal(t, "applying migration 1_create_users\n", out.String())

		out.Reset()
		assert.NoError(t, run(ctx, []string{"-dsn", dsn, "-dir", dir, "version"}, &out))
		assert.Equal(t, "1\n", out.String())
	})
	t.Run("create works without -dsn", func(t *testing.T) {
		dir := t.TempDir()
		var out bytes.Buffer
		assert.NoError(t, run(ctx, []string{"-dir", dir, "create", "add_users"}, &out))
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Len(t, entries, 2)
		assert.Regexp(t, regexp.MustCompile(`^\d{14}_add_users\.down\.sql$`), entries[0].Name())
		assert.Regexp(t, regexp.MustCompile(`^\d{14}_add_users\.up\.sql$`), entries[1].Name())
	})
	t.Run("requires -dsn except for create", func(t *testing.T) {
		var out bytes.Buffer
		assert.EqualError(t, run(ctx, []string{"up"}, &out), "-dsn is required")
	})
	t.Run("rejects an unknown flag", func(t *testing.T) {
		var out bytes.Buffer
		err := run(ctx, []string{"-unknown"}, &out)
		assert.Error(t, err)
		assert.Contains(t, out.String(), "Usage of migrate:")
	})
	t.Run("prints usage without a subcommand", func(t *testing.T) {
		var out bytes.Buffer
		err := run(ctx, nil, &out)
		assert.ErrorContains(t, err, "expects a subcommand")
	})
}
