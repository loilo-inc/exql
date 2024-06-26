package exql_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/loilo-inc/exql/v2"

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
