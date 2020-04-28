package exql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTags(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		tags, err := ParseTags("a:1;b:2;c:3")
		assert.Nil(t, err)
		assert.Equal(t, len(tags), 3)
		assert.Equal(t, "1", tags["a"])
		assert.Equal(t, "2", tags["b"])
		assert.Equal(t, "3", tags["c"])
	})
	t.Run("key only", func(t *testing.T) {
		tags, err := ParseTags("a;b;c;")
		assert.Nil(t, err)
		assert.Equal(t, len(tags), 3)
		assert.Equal(t, "", tags["a"])
		assert.Equal(t, "", tags["b"])
		assert.Equal(t, "", tags["c"])
	})
	t.Run("should return error for duplicate tag", func(t *testing.T) {
		tags, err := ParseTags("a:1;a:2")
		assert.Nil(t, tags)
		assert.Errorf(t, err, "duplicate tag: a")
	})
	t.Run("should return error if tag is empty", func(t *testing.T) {
		assertEmpty := func(s string) {
			tags, err := ParseTags(s)
			assert.Nil(t, tags)
			assert.Errorf(t, err, "empty tags")
		}
		assertEmpty("")
		assertEmpty(";:;")
		assertEmpty(":::")
		assertEmpty(";;;")
	})
}
