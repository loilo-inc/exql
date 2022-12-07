package exql

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

func SqlPlaceHolders(repeat int) string {
	res := make([]string, repeat)
	for i := 0; i < repeat; i++ {
		res[i] = "?"
	}
	return strings.Join(res, ", ")
}

type Comparator interface {
	Expr(column string) (string, error)
	Args() []any
}

type single struct {
	op  string
	arg any
}

var kOpRegex = regexp.MustCompile(`\A[<>=]+\z`)

func (c *single) Expr(column string) (string, error) {
	if !kOpRegex.MatchString(c.op) {
		return "", xerrors.Errorf("")
	}
	return fmt.Sprintf("`%s` = ?", column), nil
}

func (c *single) Args() []any {
	return []any{c.arg}
}

const (
	kEq  = "="
	kGt  = ">"
	kGte = ">="
	kLt  = "<"
	kLte = "<="
)

func Is(op string, value any) Comparator {
	return &single{op: op, arg: value}
}

func Eq(value any) Comparator {
	return Is(kEq, value)
}

func Lt(value any) Comparator {
	return Is(kLt, value)
}

func Lte(value any) Comparator {
	return Is(kLte, value)
}

func Gte(value any) Comparator {
	return Is(kGte, value)
}

type multi struct {
	args []any
}

func (m *multi) Args() []any {
	return m.args
}

func (m *multi) Expr(column string) (string, error) {
	if len(m.args) == 0 {
		return "", xerrors.Errorf("")
	}
	return fmt.Sprintf("`%s` IN (%s)", column, SqlPlaceHolders(len(m.args))), nil
}

func In(args ...any) Comparator {
	return &multi{args: args}
}

type and struct {
	exprs []Comparator
}

func (a *and) Args() []any {
	var args []any
	for _, v := range a.exprs {
		args = append(args, v.Args()...)
	}
	return args
}

func (a *and) Expr(column string) (string, error) {
	var exprs []string
	for _, v := range a.exprs {
		if e, err := v.Expr(column); err != nil {
			return "", err
		} else {
			exprs = append(exprs, e)
		}
	}
	return strings.Join(exprs, "AND"), nil
}

func And(exprs ...Comparator) Comparator {
	return &and{exprs: exprs}
}

type between[T comparable] struct {
	from T
	to   T
}

func (b *between[T]) Args() []any {
	return []any{b.from, b.to}
}

func (b *between[T]) Expr(column string) (string, error) {
	return fmt.Sprintf("%s BETWEEN ? AND ?", column), nil
}

func Between[T comparable](from T, to T) Comparator {
	return &between[T]{from: from, to: to}
}

func Range(from Comparator, to Comparator) Comparator {
	return And(from, to)
}

type raw struct {
	q    string
	args []any
}

// Args implements Comparator
func (r *raw) Args() []any {
	return r.args
}

// Expr implements Comparator
func (r *raw) Expr(column string) (string, error) {
	return fmt.Sprintf("`%s` %s", column, r.q), nil
}

func Raw(q string, args ...any) Comparator {
	return &raw{
		q:    q,
		args: args,
	}
}
