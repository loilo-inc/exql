package exql

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGenerator_Generate(t *testing.T) {
	for version, db := range map[string]DB{
		"mysql8": testDb(),
	} {
		t.Run(version, func(t *testing.T) {
			g := NewGenerator(db.DB())
			checkFiles := func(dir string, elements []string) {
				entries, err := os.ReadDir(dir)
				assert.NoError(t, err)
				var files []string
				for _, e := range entries {
					files = append(files, e.Name())
				}
				assert.ElementsMatch(t, files, elements)
			}
			t.Run("basic", func(t *testing.T) {
				dir := t.TempDir()
				err := g.Generate(&GenerateOptions{
					OutDir:  dir,
					Package: "dist",
				})
				assert.NoError(t, err)
				checkFiles(dir, []string{"users.go", "user_groups.go", "user_login_histories.go", "group_users.go", "fields.go"})
			})
			t.Run("exclude", func(t *testing.T) {
				dir := t.TempDir()
				err := g.Generate(&GenerateOptions{
					OutDir:  dir,
					Package: "dist",
					Exclude: []string{"fields"},
				})
				assert.NoError(t, err)
				checkFiles(dir, []string{"users.go", "user_groups.go", "user_login_histories.go", "group_users.go"})
			})
			t.Run("creates output dir with permission 0750", func(t *testing.T) {
				dir := filepath.Join(t.TempDir(), "output")
				err := g.Generate(&GenerateOptions{
					OutDir:  dir,
					Package: "dist",
				})
				assert.NoError(t, err)
				info, err := os.Stat(dir)
				assert.NoError(t, err)
				assert.Equal(t, os.FileMode(0750), info.Mode().Perm())
			})
			t.Run("writes files with permission 0640", func(t *testing.T) {
				dir := t.TempDir()
				err := g.Generate(&GenerateOptions{
					OutDir:  dir,
					Package: "dist",
				})
				assert.NoError(t, err)
				entries, err := os.ReadDir(dir)
				assert.NoError(t, err)
				for _, e := range entries {
					info, err := os.Stat(filepath.Join(dir, e.Name()))
					assert.NoError(t, err)
					assert.Equal(t, os.FileMode(0640), info.Mode().Perm(), "file: %s", e.Name())
				}
			})
		})
	}

	t.Run("should return error when rows.Error() return error", func(t *testing.T) {
		mockDb, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDb.Close()

		mock.ExpectQuery(`show tables`).WillReturnRows(
			sqlmock.NewRows([]string{"tables"}).
				AddRow("users").
				RowError(0, fmt.Errorf("err")))

		dir := t.TempDir()
		assert.EqualError(t, NewGenerator(mockDb).
			Generate(&GenerateOptions{
				OutDir:  dir,
				Package: "dist",
			}), "err")
	})

	t.Run("should propagate ParseTable error", func(t *testing.T) {
		mockDb, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDb.Close()

		mock.ExpectQuery(`show tables`).WillReturnRows(
			sqlmock.NewRows([]string{"tables"}).AddRow("users"))
		mock.ExpectQuery("show columns from `users`").WillReturnError(fmt.Errorf("columns err"))

		dir := t.TempDir()
		assert.EqualError(t, NewGenerator(mockDb).
			Generate(&GenerateOptions{
				OutDir:  dir,
				Package: "dist",
			}), "columns err")
	})
}
