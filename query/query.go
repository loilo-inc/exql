//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

import (
	"fmt"

	"golang.org/x/xerrors"
)

var errEmptyTable = xerrors.Errorf("empty table")
var errEmyptValues = xerrors.Errorf("empty values")
var errEmptyWhereClause = xerrors.Errorf("empty where clause")

type Query interface {
	Query() (string, []any, error)
}

type query struct {
	query string
	args  []any
	err   error
}

func (q *query) Query() (string, []any, error) {
	if q.err != nil {
		return "", nil, q.err
	}
	return q.query, q.args, nil
}

func NewQuery(q string, args ...any) Query {
	return &query{query: q, args: args}
}

type fmtQuery struct {
	fmt string
	qs  []Query
}

func (f *fmtQuery) Query() (string, []any, error) {
	var fmtArgs []any
	var sqlArgs []any
	for _, v := range f.qs {
		stmt, args, err := v.Query()
		if err != nil {
			return "", nil, err
		}
		fmtArgs = append(fmtArgs, stmt)
		sqlArgs = append(sqlArgs, args...)
	}
	return fmt.Sprintf(f.fmt, fmtArgs...), sqlArgs, nil
}

func Where(q string, args ...interface{}) Condition {
	return &chain{qs: []Query{NewQuery(q, args...)}}
}

type Condition interface {
	Query
	And(other ...Condition) Condition
	Or(other ...Condition) Condition
}

type chain struct {
	op string
	qs []Query
}

func (c *chain) And(other ...Condition) Condition {
	return c.join(" AND ", other...)
}

func (c *chain) Or(other ...Condition) Condition {
	return c.join(" OR ", other...)
}

func (c *chain) join(sep string, other ...Condition) Condition {
	list := []Query{c}
	list = append(list, c.qs...)
	return &chain{
		op: sep,
		qs: list,
	}
}

func (c *chain) Query() (string, []any, error) {
	b := NewBuilder()
	for _, v := range c.qs {
		b.Add(v)
	}
	return b.Join(c.op).Query()
}

func Q(q string, args ...any) Query {
	return NewQuery(q, args...)
}

func Cols(cols []string) Query {
	return &query{
		query: backQuoteAndJoin(cols...),
	}
}

func Vals[T any](vals []T) Query {
	var args []any
	for _, v := range vals {
		args = append(args, v)
	}
	return &query{
		query: fmt.Sprintf("(%s)", Placeholders(len(vals))),
		args:  args,
	}
}

func Set(m map[string]any) Query {
	b := NewBuilder()
	it := NewKeyIterator(m)
	for i := 0; i < it.Size(); i++ {
		k, v := it.Get(i)
		b.Sprintf("`%s` = ?", k).Args(v)
	}
	return b.Join(",")
}
