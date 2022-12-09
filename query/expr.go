//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

type Expr interface {
	Expr(column string) (string, []any, error)
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

func In[T any](args ...T) Expr {
	var arr = make([]any, len(args))
	for i, v := range args {
		arr[i] = v
	}
	return &raw{
		q:    fmt.Sprintf("IN (%s)", SqlPlaceHolders(len(args))),
		args: arr,
	}
}

type and struct {
	exprs []Expr
}

func (a *and) Expr(column string) (string, []any, error) {
	var exprs []string
	var args []any
	for _, v := range a.exprs {
		if e, a, err := v.Expr(column); err != nil {
			return "", nil, err
		} else {
			exprs = append(exprs, e)
			args = append(args, a...)
		}
	}
	return strings.Join(exprs, " AND "), args, nil
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

func (r *raw) Expr(column string) (string, []any, error) {
	if emptyPat.MatchString(column) || emptyPat.MatchString(r.q) {
		return "", nil, ErrDangerousExpr
	}
	return fmt.Sprintf("`%s` %s", column, r.q), r.args, nil
}

var ErrDangerousExpr = xerrors.Errorf("DANGER: empty where clause")
var emptyPat = regexp.MustCompile(`\A[\s\t\n]*\z`)

func GuardDangerousQuery(s string) error {
	if emptyPat.MatchString(s) {
		return ErrDangerousExpr
	}
	return nil
}

func Raw(q string, args ...any) Expr {
	return &raw{
		q:    q,
		args: args,
	}
}
