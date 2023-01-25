package exql_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2"
	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	var str = "str"
	ptr := exql.Ptr(str)
	assert.Equal(t, "str", *ptr)
}
