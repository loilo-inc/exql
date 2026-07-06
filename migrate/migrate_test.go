package migrate

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		fsys := fstest.MapFS{
			"2_add_age.up.sql":      {Data: []byte("alter table users add age int")},
			"2_add_age.down.sql":    {Data: []byte("alter table users drop age")},
			"1_create_users.up.sql": {Data: []byte("create table users (id int)")},
			"10_no_down.up.sql":     {Data: []byte("create table t10 (id int)")},
			"README.md":             {Data: []byte("ignored")},
		}
		migrations, err := Load(fsys)
		assert.NoError(t, err)
		assert.Equal(t, []*Migration{
			{Version: 1, Name: "create_users", UpStmt: "create table users (id int)"},
			{Version: 2, Name: "add_age", UpStmt: "alter table users add age int", DownStmt: "alter table users drop age", HasDown: true},
			{Version: 10, Name: "no_down", UpStmt: "create table t10 (id int)"},
		}, migrations)
	})
	t.Run("empty files are loaded as no-op migrations", func(t *testing.T) {
		fsys := fstest.MapFS{
			"1_a.up.sql":   {},
			"1_a.down.sql": {},
		}
		migrations, err := Load(fsys)
		assert.NoError(t, err)
		assert.Equal(t, []*Migration{
			{Version: 1, Name: "a", HasDown: true},
		}, migrations)
	})
	t.Run("missing up migration", func(t *testing.T) {
		fsys := fstest.MapFS{
			"1_a.down.sql": {Data: []byte("x")},
		}
		_, err := Load(fsys)
		assert.EqualError(t, err, "missing up migration for 1_a")
	})
	t.Run("duplicate version with different names", func(t *testing.T) {
		fsys := fstest.MapFS{
			"1_a.up.sql": {Data: []byte("x")},
			"1_b.up.sql": {Data: []byte("y")},
		}
		_, err := Load(fsys)
		assert.ErrorContains(t, err, "duplicate")
		assert.ErrorContains(t, err, "version 1")
	})
	t.Run("duplicate up migration", func(t *testing.T) {
		fsys := fstest.MapFS{
			"01_a.up.sql": {Data: []byte("x")},
			"1_a.up.sql":  {Data: []byte("y")},
		}
		_, err := Load(fsys)
		assert.EqualError(t, err, "duplicate up migration for version 1")
	})
	t.Run("duplicate down migration", func(t *testing.T) {
		fsys := fstest.MapFS{
			"1_a.up.sql":    {Data: []byte("x")},
			"01_a.down.sql": {Data: []byte("y")},
			"1_a.down.sql":  {Data: []byte("z")},
		}
		_, err := Load(fsys)
		assert.EqualError(t, err, "duplicate down migration for version 1")
	})
	t.Run("version out of range", func(t *testing.T) {
		fsys := fstest.MapFS{
			"99999999999999999999_a.up.sql": {Data: []byte("x")},
		}
		_, err := Load(fsys)
		assert.ErrorContains(t, err, "invalid migration version")
	})
	t.Run("empty", func(t *testing.T) {
		migrations, err := Load(fstest.MapFS{})
		assert.NoError(t, err)
		assert.Empty(t, migrations)
	})
}
