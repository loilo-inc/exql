//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

import (
	"fmt"
	"regexp"

	"golang.org/x/xerrors"
)

type Predicate interface {
	Predicate(column string) (string, []any, error)
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

func is(op op, value any) Predicate {
	return &pred{
		q:    fmt.Sprintf("%s ?", op),
		args: []any{value},
	}
}

func Eq(value any) Predicate {
	return is(kEq, value)
}

func NotEq(value any) Predicate {
	return is(kNotEq, value)
}

func IsNull() Predicate {
	return &pred{q: kIsNull}
}

func IsNotNull() Predicate {
	return &pred{q: kIsNotNull}
}

func Like(expr string) Predicate {
	return is(kLike, expr)
}

func Lt(value any) Predicate {
	return is(kLt, value)
}

func Lte(value any) Predicate {
	return is(kLte, value)
}
func Gt(value any) Predicate {
	return is(kGt, value)
}
func Gte(value any) Predicate {
	return is(kGte, value)
}

func In[T any](args ...T) Predicate {
	var arr = make([]any, len(args))
	for i, v := range args {
		arr[i] = v
	}
	return &pred{
		q:    fmt.Sprintf("IN (%s)", Placeholders(len(args))),
		args: arr,
	}
}

func Between[T comparable](from T, to T) Predicate {
	return &pred{
		q:    "BETWEEN ? AND ?",
		args: []any{from, to},
	}
}

type pred struct {
	q    string
	args []any
}

func (r *pred) Predicate(column string) (string, []any, error) {
	if emptyPat.MatchString(column) || emptyPat.MatchString(r.q) {
		return "", nil, errEmptyPred
	}
	return fmt.Sprintf("`%s` %s", column, r.q), r.args, nil
}

var errEmptyPred = xerrors.Errorf("DANGER: empty predicate")
var emptyPat = regexp.MustCompile(`\A[\s\t\n]*\z`)

type multiPred struct {
	op    string
	preds []Predicate
}

func (m *multiPred) Predicate(column string) (string, []any, error) {
	var preds []string
	var args []any
	for _, v := range m.preds {
		if e, a, err := v.Predicate(column); err != nil {
			return "", nil, err
		} else {
			preds = append(preds, e)
			args = append(args, a...)
		}
	}
	if ret, err := concatQueries(m.op, preds); err != nil {
		return "", nil, err
	} else {
		return ret, args, nil
	}
}

func PredicateAnd(list ...Predicate) Predicate {
	return concatPred(kAnd, list...)
}

func PredicateOr(list ...Predicate) Predicate {
	return concatPred(kOr, list...)
}

func concatPred(op string, list ...Predicate) Predicate {
	return &multiPred{op: op, preds: list}
}

func Raw(q string, args ...any) Predicate {
	return &pred{q: q, args: args}
}
