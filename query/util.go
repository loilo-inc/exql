package query

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type keyIterator[T any] struct {
	keys   []string
	values []T
}

type KeyIterator[T any] interface {
	Get(i int) (string, T)
	Keys() []string
	Values() []T
	Size() int
	Map() map[string]T
}

func NewKeyIterator[T any](data map[string]T) KeyIterator[T] {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) < 0
	})
	var values []T
	for _, v := range keys {
		values = append(values, data[v])
	}
	return &keyIterator[T]{keys: keys, values: values}
}

func (e *keyIterator[T]) Get(i int) (string, T) {
	k := e.keys[i]
	v := e.values[i]
	return k, v
}

func (e *keyIterator[T]) Size() int {
	return len(e.keys)
}

func (k *keyIterator[T]) Keys() []string {
	return k.keys
}

func (k *keyIterator[T]) Values() []T {
	return k.values
}

func (k *keyIterator[T]) Map() map[string]T {
	res := map[string]T{}
	for i := 0; i < k.Size(); i++ {
		res[k.keys[i]] = k.values[i]
	}
	return res
}

// Placeholders makes n-th repeats of Go's SQL placeholder(?),
// joining them by comma(,).
//
// Example:
//
//	Placeholder(2) // ?,?
func Placeholders(repeat int) string {
	res := make([]string, repeat)
	for i := 0; i < repeat; i++ {
		res[i] = "?"
	}
	return strings.Join(res, ",")
}

// QuoteColumns quotes each string and joins them by comma(,).
func QuoteColumns(str ...string) string {
	var result []string
	for _, v := range str {
		result = append(result, QuoteColumn(v))
	}
	return strings.Join(result, ",")
}

func guardQuery(q string) error {
	if q == "" {
		return errors.New("DANGER: empty query")
	}
	return nil
}

// QuoteColumn surrounds SQL identifiers with backquote,
// keeping some meta-characters "*", ".", "`" intact.
//
// Example:
//
//	QuoteColumn("users.id") // `users`.`id`
//	QuoteColumn("users.*") // `users`.*
func QuoteColumn(col string) string {
	var sb strings.Builder
	var start = 0
	var end = len(col)
	for i := 0; i < end; i++ {
		char := col[i]
		if char == '.' || char == '*' || char == '`' {
			if start != i {
				sb.WriteString(fmt.Sprintf("`%s`", col[start:i]))
			}
			if char != '`' {
				sb.WriteByte(char)
			}
			start = i + 1
		}
	}
	if start < end {
		sb.WriteString(fmt.Sprintf("`%s`", col[start:end]))
	}
	return sb.String()
}
