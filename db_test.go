package exql_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestDb_DB(t *testing.T) {
	db := testDb()
	db.SetDB(nil)
	assert.Nil(t, db.DB())
}

func TestNewDB(t *testing.T) {
	d := testSqlDB()
	db := exql.NewDB(d)
	assert.Equal(t, d, db.DB())
}

func TestOpen(t *testing.T) {
	t.Run("should call OpenContext", func(t *testing.T) {
		d, err := exql.Open(&exql.OpenOptions{
			Url: test.DbUrl,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, d)
	})
}

func TestOpenContext(t *testing.T) {
	t.Run("should return error when url is empty", func(t *testing.T) {
		_, err := exql.OpenContext(context.TODO(), &exql.OpenOptions{
			Url: "",
		})
		assert.EqualError(t, err, "opts.Url is required")
	})
	t.Run("with custom opener", func(t *testing.T) {
		var called bool
		_, err := exql.OpenContext(context.TODO(), &exql.OpenOptions{
			Url: test.DbUrl,
			OpenFunc: func(driverName string, url string) (*sql.DB, error) {
				called = true
				return sql.Open(driverName, url)
			},
		})
		assert.NoError(t, err)
		assert.True(t, called)
	})
}
