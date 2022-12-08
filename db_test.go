package exql

import (
	"database/sql"
	"sync"
	"testing"

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

func TestDb_SetDB(t *testing.T) {
	d := testSqlDB()
	t.Run("should also set saver's ex", func(t *testing.T) {
		_db := &db{
			db:    d,
			s:     &saver{ex: d},
			mutex: sync.Mutex{},
		}
		var nilPtr *sql.DB
		_db.SetDB(nilPtr)
		assert.Equal(t, nilPtr, _db.db)
		assert.Equal(t, nilPtr, _db.s.ex)
	})
}
