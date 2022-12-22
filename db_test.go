package exql_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2"
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
