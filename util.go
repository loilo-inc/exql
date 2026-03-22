package exql

import "sync"

// Ptr returns the pointer of the argument.
func Ptr[T any](t T) *T {
	return &t
}

type syncMap[K comparable, V any] struct {
	m sync.Map
}

func (m *syncMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}

func (m *syncMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *syncMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
