//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

import (
	"fmt"
	"regexp"
	"strings"
)

type Query interface {
	Query() (string, []any, error)
}

type query struct {
	query string
	args  []any
	err   error
}

func errQuery(err error) Query {
	return &query{err: err}
}

func (f *query) Query() (sqlStmt string, sqlArgs []any, resErr error) {
	if f.err != nil {
		resErr = f.err
		return
	}
	str := f.query
	args := f.args
	sb := &strings.Builder{}
	var argIdx = 0
	reg := regexp.MustCompile(`:?\?`)
	for {
		match := reg.FindStringIndex(str)
		if match == nil {
			break
		}
		if argIdx == len(args) {
			resErr = fmt.Errorf("missing argument at %d", argIdx)
			return
		}
		mStart := match[0]
		mEnd := match[1]
		if mEnd-mStart == 2 {
			// :?
			if q, ok := args[argIdx].(Query); !ok {
				resErr = fmt.Errorf("unexpected argument type for :? placeholder at %d", argIdx)
				return
			} else if stmt, vals, err := q.Query(); err != nil {
				resErr = err
				return
			} else {
				pre := str[:mStart]
				sb.WriteString(pre)
				sb.WriteString(stmt)
				sqlArgs = append(sqlArgs, vals...)
			}
		} else {
			// ?
			sb.WriteString(str[:mEnd])
			sqlArgs = append(sqlArgs, args[argIdx])
		}
		str = str[mEnd:]
		argIdx += 1
	}
	if len(args) != argIdx {
		resErr = fmt.Errorf("arguments count mismatch: found %d, got %d", argIdx, len(args))
		return
	}
	if len(str) > 0 {
		sb.WriteString(str)
	}
	sqlStmt = sb.String()
	if resErr = guardQuery(sqlStmt); resErr != nil {
		return
	}
	return sqlStmt, sqlArgs, nil
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
		if s, v, err := v.Query(); err != nil {
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

func New(q string, args ...any) Query {
	return NewBuilder().Query(q, args...).Build()
}

func Q(q string, args ...any) Query {
	return &query{
		query: q,
		args:  args,
	}
}

func Cols(cols ...string) Query {
	if len(cols) == 0 {
		return errQuery(fmt.Errorf("empty columns"))
	}
	return &query{
		query: backQuoteAndJoin(cols...),
	}
}

func V(a ...any) Query {
	return &query{
		query: Placeholders(len(a)),
		args:  a,
	}
}

func Vals[T any](vals []T) Query {
	if len(vals) == 0 {
		return errQuery(fmt.Errorf("empty values"))
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
		return errQuery(fmt.Errorf("empty values for set clause"))
	}
	b := NewBuilder()
	it := NewKeyIterator(m)
	for i := 0; i < it.Size(); i++ {
		k, v := it.Get(i)
		b.Query("`:?` = ?", Q(k), v)
	}
	return b.Join(",")
}
