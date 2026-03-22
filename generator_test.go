package exql_test

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/loilo-inc/exql/v3"

	"github.com/stretchr/testify/assert"
)

func TestGenerator_Generate(t *testing.T) {
	for version, db := range map[string]exql.DB{
		"mysql8": testDb(),
	} {
		t.Run(version, func(t *testing.T) {
			g := exql.NewGenerator(db.DB())
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
				err := g.Generate(&exql.GenerateOptions{
					OutDir:  dir,
					Package: "dist",
				})
				assert.NoError(t, err)
				checkFiles(dir, []string{"users.go", "user_groups.go", "user_login_histories.go", "group_users.go", "fields.go"})
			})
			t.Run("exclude", func(t *testing.T) {
				dir := t.TempDir()
				err := g.Generate(&exql.GenerateOptions{
					OutDir:  dir,
					Package: "dist",
					Exclude: []string{"fields"},
				})
				assert.NoError(t, err)
				checkFiles(dir, []string{"users.go", "user_groups.go", "user_login_histories.go", "group_users.go"})
			})

			t.Run("should return error when rows.Error() return error", func(t *testing.T) {
				mockDb, mock, err := sqlmock.New()
				assert.NoError(t, err)
				defer mockDb.Close()

				mock.ExpectQuery(`show tables`).WillReturnRows(
					sqlmock.NewRows([]string{"tables"}).
						AddRow("users").
						RowError(0, fmt.Errorf("err")))

				dir := t.TempDir()
				assert.EqualError(t, exql.NewGenerator(mockDb).
					Generate(&exql.GenerateOptions{
						OutDir:  dir,
						Package: "dist",
					}), "err")
			})
		})
	}
}

func TestGenerator_Generate_formatsGeneratedCode(t *testing.T) {
	mockDb, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDb.Close()

	mock.ExpectQuery(`show tables`).WillReturnRows(
		sqlmock.NewRows([]string{"tables"}).
			AddRow("user_profiles"),
	)
	mock.ExpectQuery(`show columns from user_profiles`).WillReturnRows(
		sqlmock.NewRows([]string{"field", "type", "null", "key", "default", "extra"}).
			AddRow("id", "int(11)", "NO", "PRI", nil, "auto_increment").
			AddRow("name", "varchar(255)", "NO", "", nil, "").
			AddRow("created_at", "datetime", "NO", "", nil, "").
			AddRow("metadata", "json", "YES", "", nil, ""),
	)

	dir := t.TempDir()
	err = exql.NewGenerator(mockDb).Generate(&exql.GenerateOptions{
		OutDir:  dir,
		Package: "dist",
	})
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "user_profiles.go"))
	assert.NoError(t, err)

	formatted, err := format.Source(content)
	assert.NoError(t, err)
	assert.Equal(t, string(formatted), string(content))
	assert.NoError(t, mock.ExpectationsWereMet())
}
