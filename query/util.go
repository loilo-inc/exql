package query

import (
	"fmt"
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
	Map() map[string]any
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

func (k *keyIterator) Map() map[string]any {
	res := map[string]any{}
	for i := 0; i < k.Size(); i++ {
		res[k.keys[i]] = k.values[i]
	}
	return res
}

func Placeholders(repeat int) string {
	res := make([]string, repeat)
	for i := 0; i < repeat; i++ {
		res[i] = "?"
	}
	return strings.Join(res, ",")
}

func backQuoteAndJoin(str ...string) string {
	var result []string
	for _, v := range str {
		result = append(result, fmt.Sprintf("`%s`", v))
	}
	return strings.Join(result, ",")
}
