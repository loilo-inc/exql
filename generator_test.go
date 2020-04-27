package exql

import (
	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	db, err := Open(&OpenOptions{
		Url: "root:@tcp(127.0.0.1:3326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	if err != nil {
		panic(err)
		return
	}
	err = db.Generate(&GenerateOptions{
		OutDir:  "dist",
		Package: "dist",
	})
	if err != nil {
		log.Errorf(err.Error())
	}
	assert.Nil(t, err)
}
