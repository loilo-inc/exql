package query_test

import (
	"testing"

	. "github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
)

func TestKeyIteretor(t *testing.T) {
	it := NewKeyIterator(map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
	})
	assert.Equal(t, it.Size(), 3)
	assert.ElementsMatch(t, it.Keys(), []string{"a", "b", "c"})
	assert.ElementsMatch(t, it.Values(), []any{1, 2, 3})
	for i := 0; i < it.Size(); i++ {
		k, v := it.Get(i)
		assert.Equal(t, it.Keys()[i], k)
		assert.Equal(t, it.Values()[i], v)
	}
}

func TestSqlPraceholder(t *testing.T) {
	assert.Equal(t, "", Placeholders(0))
	assert.Equal(t, "?", Placeholders(1))
	assert.Equal(t, "?,?,?", Placeholders(3))
}
