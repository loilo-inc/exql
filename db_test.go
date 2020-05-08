package exql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDb_DB(t *testing.T) {
	db := testDb()
	db.SetDB(nil)
	assert.Nil(t, db.DB())
}
