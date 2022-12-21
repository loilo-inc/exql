//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

import (
	"fmt"
	"regexp"

	"golang.org/x/xerrors"
)

type Expression interface {
	Expression(column string) (string, []any, error)
}

type op = string

const (
	kEq        op = "="
	kNotEq     op = "<>"
	kGt        op = ">"
	kGte       op = ">="
	kLt        op = "<"
	kLte       op = "<="
	kIsNull    op = "IS NULL"
	kIsNotNull op = "IS NOT NULL"
	kLike      op = "LIKE"
)

func is(op op, value any) Expression {
	return &raw{
		q:    fmt.Sprintf("%s ?", op),
		args: []any{value},
	}
}

func Eq(value any) Expression {
	return is(kEq, value)
}

func NotEq(value any) Expression {
	return is(kNotEq, value)
}

func IsNull() Expression {
	return &raw{q: kIsNull}
}

func IsNotNull() Expression {
	return &raw{q: kIsNotNull}
}

func Like(Expression string) Expression {
	return is(kLike, Expression)
}

func Lt(value any) Expression {
	return is(kLt, value)
}

func Lte(value any) Expression {
	return is(kLte, value)
}
func Gt(value any) Expression {
	return is(kGt, value)
}
func Gte(value any) Expression {
	return is(kGte, value)
}

func In[T any](args ...T) Expression {
	var arr = make([]any, len(args))
	for i, v := range args {
		arr[i] = v
	}
	return &raw{
		q:    fmt.Sprintf("IN (%s)", Placeholders(len(args))),
		args: arr,
	}
}

func Between[T comparable](from T, to T) Expression {
	return &raw{
		q:    "BETWEEN ? AND ?",
		args: []any{from, to},
	}
}

type raw struct {
	q    string
	args []any
}

func (r *raw) Expression(column string) (string, []any, error) {
	if emptyPat.MatchString(column) || emptyPat.MatchString(r.q) {
		return "", nil, errEmptyExpr
	}
	return fmt.Sprintf("`%s` %s", column, r.q), r.args, nil
}

var errEmptyExpr = xerrors.Errorf("DANGER: empty expression")
var emptyPat = regexp.MustCompile(`\A[\s\t\n]*\z`)

type multiExpression struct {
	op          string
	Expressions []Expression
}

func (m *multiExpression) Expression(column string) (string, []any, error) {
	var Expressions []string
	var args []any
	for _, v := range m.Expressions {
		if e, a, err := v.Expression(column); err != nil {
			return "", nil, err
		} else {
			Expressions = append(Expressions, e)
			args = append(args, a...)
		}
	}
	if ret, err := concatQueries(m.op, Expressions); err != nil {
		return "", nil, err
	} else {
		return ret, args, nil
	}
}

func ExpressionAnd(list ...Expression) Expression {
	return concatExpression(kAnd, list...)
}

func ExpressionOr(list ...Expression) Expression {
	return concatExpression(kOr, list...)
}

func concatExpression(op string, list ...Expression) Expression {
	return &multiExpression{op: op, Expressions: list}
}

func Raw(q string, args ...any) Expression {
	return &raw{
		q:    q,
		args: args,
	}
}
