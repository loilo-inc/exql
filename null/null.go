package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Nuller interface {
	sql.Scanner
	driver.Valuer
	json.Marshaler
}

type Null[T any] struct {
	sql.Null[T]
}

func New[T any](v T) Null[T] {
	return Null[T]{
		Null: sql.Null[T]{V: v, Valid: true},
	}
}

var _ Nuller = (*Null[any])(nil)

// MarshalJSON implements json.Marshaler.
func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.V)
}

type Uint64 = Null[uint64]
type Int64 = Null[int64]
type Float64 = Null[float64]
type Float32 = Null[float32]
type Time = Null[time.Time]
type String = Null[string]
type Bytes = Null[[]byte]
type JSON = Null[json.RawMessage]
