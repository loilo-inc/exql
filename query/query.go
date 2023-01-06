//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

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
	} else if err := guardQuery(q.query); err != nil {
		return "", nil, err
	}
	return q.query, q.args, nil
}

func errQuery(err error) Query {
	return &query{err: err}
}

type fmtQuery struct {
	fmt string
	qs  []Query
}

func (f *fmtQuery) Query() (string, []any, error) {
	var fmtArgs []any
	var sqlArgs []any
	for _, v := range f.qs {
		if stmt, args, err := v.Query(); err != nil {
			return "", nil, err
		} else {
			fmtArgs = append(fmtArgs, stmt)
			sqlArgs = append(sqlArgs, args...)
		}
	}
	stmt := fmt.Sprintf(f.fmt, fmtArgs...)
	if err := guardQuery(stmt); err != nil {
		return "", nil, err
	}
	return stmt, sqlArgs, nil
}

type Condition interface {
	Query
	And(str string, args ...any)
	Or(str string, args ...any)
	AndCond(other Condition)
	OrCond(other Condition)
}

func Cond(str string, args ...any) Condition {
	return CondFrom(Q(str, args...))
}

func CondFrom(q ...Query) Condition {
	base := &chain{
		joiner: " ",
		list:   q,
	}
	return &cond{base: base}
}

type cond struct {
	base *chain
}

func (c *cond) And(str string, args ...any) {
	c.append("AND", Q(str, args...))
}

func (c *cond) Or(str string, args ...any) {
	c.append("OR", Q(str, args...))
}

func (c *cond) AndCond(other Condition) {
	c.append("AND", other)
}

func (c *cond) OrCond(other Condition) {
	c.append("OR", other)
}

func (c *cond) Query() (string, []any, error) {
	return c.base.Query()
}

func (c *cond) append(sep string, other ...Query) {
	joiner := Q(sep)
	for _, v := range other {
		if len(c.base.list) == 0 {
			c.base.append(v)
		} else {
			c.base.append(joiner, v)
		}
	}
}

type chain struct {
	joiner string
	list   []Query
}

func (c *chain) append(other ...Query) {
	c.list = append(c.list, other...)
}

func (c *chain) Query() (string, []any, error) {
	var strs []string
	var args []any
	for _, v := range c.list {
		if s, v, err := v.Query(); err == errArgsOnly {
			args = append(args, v...)
		} else if err != nil {
			return "", nil, err
		} else {
			strs = append(strs, s)
			args = append(args, v...)
		}
	}
	stmt := strings.Join(strs, c.joiner)
	if err := guardQuery(stmt); err != nil {
		return "", nil, err
	}
	return stmt, args, nil
}

type argsOnly struct {
	args []any
}

var errArgsOnly = xerrors.New("argsOnly does't buid query")

func (a *argsOnly) Query() (string, []any, error) {
	return "", a.args, errArgsOnly
}

func Qprintf(q string, qs ...Query) Query {
	return NewBuilder().Qprintf(q, qs...).Build()
}

func Q(q string, args ...any) Query {
	return &query{
		query: q,
		args:  args,
	}
}

func Cols(cols []string) Query {
	if len(cols) == 0 {
		return errQuery(xerrors.Errorf("empty columns"))
	}
	return &query{
		query: backQuoteAndJoin(cols...),
	}
}

func Val(a any) Query {
	return &query{
		query: "?",
		args:  []any{a},
	}
}

func Vals[T any](vals []T) Query {
	if len(vals) == 0 {
		return errQuery(xerrors.Errorf("empty values"))
	}
	var args []any
	for _, v := range vals {
		args = append(args, v)
	}
	return &query{
		query: Placeholders(len(vals)),
		args:  args,
	}
}

func Set(m map[string]any) Query {
	if len(m) == 0 {
		return errQuery(xerrors.Errorf("empty values for set clause"))
	}
	b := NewBuilder()
	it := NewKeyIterator(m)
	for i := 0; i < it.Size(); i++ {
		k, v := it.Get(i)
		b.Sprintf("`%s` = ?", k).Args(v)
	}
	return b.Join(",")
}
