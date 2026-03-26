package mock

import "github.com/loilo-inc/exql/v3/iface"

type Rows struct {
	Cols      []string
	Values    [][]any
	Idx       int
	ColumnErr error
	ScanErr   error
	ErrErr    error
	CloseErr  error
}

var _ iface.SqlRows = (*Rows)(nil)

func (r *Rows) Columns() ([]string, error) {
	return r.Cols, r.ColumnErr
}

func (r *Rows) Next() bool {
	if r.Idx >= len(r.Values) {
		return false
	}
	r.Idx++
	return true
}

func (r *Rows) Scan(dest ...any) error {
	return r.ScanErr
}

func (r *Rows) Err() error {
	return r.ErrErr
}

func (r *Rows) Close() error {
	return r.CloseErr
}
