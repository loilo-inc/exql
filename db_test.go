package exql

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/loilo-inc/exql/v3/model"
	"github.com/stretchr/testify/assert"
)

func TestDb_DB(t *testing.T) {
	db := testDb()
	db.SetDB(nil)
	assert.Nil(t, db.DB())
}

func TestNewDB(t *testing.T) {
	d := testSqlDB()
	db := NewDB(d)
	assert.Equal(t, d, db.DB())
}

func TestDb_Close(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	mock.ExpectClose()

	db := NewDB(mockDB).(*db)
	s1, err := db.reflector.GetSchema(reflect.TypeFor[model.Users](), false)
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	s2, err := db.reflector.GetSchema(reflect.TypeFor[model.Users](), false)
	assert.NoError(t, err)
	assert.NotSame(t, s1, s2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDb_SetDB(t *testing.T) {
	db1, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db1.Close()

	db2, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db2.Close()

	d := NewDB(db1).(*db)
	s1, err := d.reflector.GetSchema(reflect.TypeFor[model.Users](), false)
	assert.NoError(t, err)

	d.SetDB(db2)

	assert.Same(t, db2, d.DB())
	assert.Same(t, db2, d.saver.ex)
	assert.Same(t, db2, d.finder.ex)

	s2, err := d.reflector.GetSchema(reflect.TypeFor[model.Users](), false)
	assert.NoError(t, err)
	assert.NotSame(t, s1, s2)
}

func TestDb_Transaction(t *testing.T) {
	t.Run("commit", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectBegin()
		mock.ExpectCommit()

		db := NewDB(mockDB)
		err = db.Transaction(func(tx Tx) error {
			assert.NotNil(t, tx.Tx())
			return nil
		})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback on callback error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		db := NewDB(mockDB)
		err = db.Transaction(func(tx Tx) error {
			assert.NotNil(t, tx.Tx())
			return assert.AnError
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestOpen(t *testing.T) {
	t.Run("should call OpenContext", func(t *testing.T) {
		d, err := Open(&OpenOptions{
			Url: dbUrl,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, d)
	})
}

func TestOpenContext(t *testing.T) {
	t.Run("should return error when url is empty", func(t *testing.T) {
		_, err := OpenContext(context.TODO(), &OpenOptions{
			Url: "",
		})
		assert.EqualError(t, err, "opts.Url is required")
	})
	t.Run("with custom opener", func(t *testing.T) {
		var called bool
		_, err := OpenContext(context.TODO(), &OpenOptions{
			Url: dbUrl,
			OpenFunc: func(driverName string, url string) (*sql.DB, error) {
				called = true
				return sql.Open(driverName, url)
			},
		})
		assert.NoError(t, err)
		assert.True(t, called)
	})
}
