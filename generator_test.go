package exql

import (
	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	db := testDb()
	g := NewGenerator(db.DB())
	err := g.Generate(&GenerateOptions{
		OutDir:  "dist",
		Package: "dist",
	})
	if err != nil {
		log.Errorf(err.Error())
	}
	assert.Nil(t, err)
}
