package exql

import (
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

type ClauseType string

const (
	ClauseTypeWhere = "where"
)

type Clause interface {
	Query() (string, error)
	Args() []interface{}
}

type clause struct {
	query string
	args  []interface{}
}

var ErrDangerWhereClause = xerrors.Errorf("DANGER: empty where clause")

var emptyPat = regexp.MustCompile(`\A[\s|\t]*\z`)

func IsSafeWhereClause(s string) bool {
	return !emptyPat.MatchString(s)
}

func (w *clause) Query() (string, error) {
	if !IsSafeWhereClause(w.query) {
		return "", ErrDangerWhereClause
	}
	return w.query, nil
}

func (w *clause) Args() []interface{} {
	return w.args
}

func Where(q string, args ...interface{}) Clause {
	return &clause{
		query: q,
		args:  args,
	}
}

type clauseEx struct {
	cond map[string]any // comaparble | Comparator
}

func (c *clauseEx) Args() []interface{} {
	var args []any
	for _, v := range c.cond {
		switch e := v.(type) {
		case Comparator:
			args = append(args, e.Args()...)
		default:
			args = append(args, e)
		}
	}
	return args
}

func (c *clauseEx) Query() (string, error) {
	var arr []string
	for k, v := range c.cond {
		var expr Comparator
		switch e := v.(type) {
		case Comparator:
			expr = e
		default:
			expr = Eq(v)
		}
		if expr, err := expr.Expr(k); err != nil {
			return "", err
		} else {
			arr = append(arr, expr)
		}
	}
	query := strings.Join(arr, " AND ")
	if !IsSafeWhereClause(query) {
		return "", ErrDangerWhereClause
	}
	return query, nil
}

func WhereEx(cond map[string]any) Clause {
	return &clauseEx{cond: cond}
}
