package exql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	var str = "str"
	ptr := Ptr(str)
	assert.Equal(t, "str", *ptr)
}
