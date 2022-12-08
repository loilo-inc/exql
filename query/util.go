package query

import (
	"sort"
	"strings"
)

type keyIterator struct {
	keys   []string
	values []any
}

type KeyIterator interface {
	Get(i int) (string, any)
	Keys() []string
	Values() []any
	Size() int
}

func NewKeyIterator(data map[string]any) KeyIterator {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) < 0
	})
	var values []any
	for _, v := range keys {
		values = append(values, data[v])
	}
	return &keyIterator{keys: keys, values: values}
}

func (e *keyIterator) Get(i int) (string, any) {
	k := e.keys[i]
	v := e.values[i]
	return k, v
}

func (e *keyIterator) Size() int {
	return len(e.keys)
}

func (k *keyIterator) Keys() []string {
	return k.keys
}

func (k *keyIterator) Values() []any {
	return k.values
}

func SqlPlaceHolders(repeat int) string {
	res := make([]string, repeat)
	for i := 0; i < repeat; i++ {
		res[i] = "?"
	}
	return strings.Join(res, ",")
}
