//go:generate go run github.com/golang/mock/mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

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
	return strings.Join(res, ",")
}

type Expr interface {
	Expr(column string) (string, error)
	Args() []any
}

type op = string

const (
	kEq        op = "="
	kNotEq     op = "!="
	kGt        op = ">"
	kGte       op = ">="
	kLt        op = "<"
	kLte       op = "<="
	kIsNull    op = "IS NULL"
	kIsNotNull op = "IS NOT NULL"
)

func is(op op, value any) Expr {
	return &raw{
		q:    fmt.Sprintf("%s ?", op),
		args: []any{value},
	}
}

func Eq(value any) Expr {
	return is(kEq, value)
}

func NotEq(value any) Expr {
	return is(kNotEq, value)
}

func IsNull() Expr {
	return &raw{q: kIsNull}
}

func IsNotNull() Expr {
	return &raw{q: kIsNotNull}
}

func Lt(value any) Expr {
	return is(kLt, value)
}

func Lte(value any) Expr {
	return is(kLte, value)
}
func Gt(value any) Expr {
	return is(kGt, value)
}
func Gte(value any) Expr {
	return is(kGte, value)
}

func In(args ...any) Expr {
	return &raw{
		q:    fmt.Sprintf("IN (%s)", SqlPlaceHolders(len(args))),
		args: args,
	}
}

type and struct {
	exprs []Expr
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
	return strings.Join(exprs, " AND "), nil
}

func And(exprs ...Expr) Expr {
	return &and{exprs: exprs}
}

func Between[T comparable](from T, to T) Expr {
	return &raw{
		q:    "BETWEEN ? AND ?",
		args: []any{from, to},
	}
}

type raw struct {
	q    string
	args []any
}

func (r *raw) Args() []any {
	return r.args
}

func (r *raw) Expr(column string) (string, error) {
	if emptyPat.MatchString(column) || emptyPat.MatchString(r.q) {
		return "", ErrDangerousExpr
	}
	return fmt.Sprintf("`%s` %s", column, r.q), nil
}

var ErrDangerousExpr = xerrors.Errorf("DANGER: empty where clause")
var emptyPat = regexp.MustCompile(`\A[\s\t\n]*\z`)

func GuardDangerousQuery(s string) (string, error) {
	if emptyPat.MatchString(s) {
		return "", ErrDangerousExpr
	}
	return s, nil
}

func Raw(q string, args ...any) Expr {
	return &raw{
		q:    q,
		args: args,
	}
}
