package exql

import (
	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	g := NewGenerator(&Options{
		OutDir:  "model",
		Url:     "root:@tcp(127.0.0.1:3326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
		Package: "model",
	})
	err := g.Generate()
	if err != nil {
		log.Errorf(err.Error())
	}
	assert.Nil(t, err)
}