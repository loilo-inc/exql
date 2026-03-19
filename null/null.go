package null

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Nuller interface {
	sql.Scanner
	driver.Valuer
	json.Marshaler
	json.Unmarshaler
}

type Null[T any] struct {
	sql.Null[T]
}

func New[T any](v T) Null[T] {
	return Null[T]{
		Null: sql.Null[T]{V: v, Valid: true},
	}
}

func FromPtr[T any](v *T) Null[T] {
	if v == nil {
		return Null[T]{}
	}
	return New(*v)
}

var _ Nuller = (*Null[any])(nil)

// MarshalJSON implements json.Marshaler.
func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.V)
}

var nullBytes = []byte("null")

// UnmarshalJSON implements json.Unmarshaler.
func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), nullBytes) {
		n.V = *new(T)
		n.Valid = false
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	n.V = v
	n.Valid = true
	return nil
}

func (n *Null[T]) Ptr() *T {
	if !n.Valid {
		return nil
	}
	return &n.V
}

type Uint64 = Null[uint64]
type Int64 = Null[int64]
type Bool = Null[bool]
type Float64 = Null[float64]
type Float32 = Null[float32]
type Time = Null[time.Time]
type String = Null[string]
type Bytes = Null[[]byte]
type JSON = Null[json.RawMessage]
