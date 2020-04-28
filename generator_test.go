package exql

import (
	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	db := testDb()
	err := db.Generate(&GenerateOptions{
		OutDir:  "dist",
		Package: "dist",
	})
	if err != nil {
		log.Errorf(err.Error())
	}
	assert.Nil(t, err)
}
