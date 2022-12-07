package exql

import (
	"fmt"
	"sort"
	"strings"

	. "github.com/loilo-inc/exql/query"
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

func (w *clause) Query() (string, error) {
	return GuardDangerousQuery(w.query)
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
	stmts []*stmt
}

func (c *clauseEx) Args() []interface{} {
	var args []any
	for _, e := range c.stmts {
		args = append(args, e.expr.Args()...)
	}
	return args
}

func (c *clauseEx) Query() (string, error) {
	var arr []string
	for _, v := range c.stmts {
		if expr, err := v.expr.Expr(v.column); err != nil {
			return "", err
		} else {
			arr = append(arr, fmt.Sprintf("(%s)", expr))
		}
	}
	query := strings.Join(arr, " AND ")
	return GuardDangerousQuery(query)
}

type stmt struct {
	column string
	expr   Expr
}

func WhereEx(cond map[string]any) Clause {
	keys := make([]string, 0, len(cond))
	for k := range cond {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) < 0
	})
	stmts := make([]*stmt, len(keys))
	for i, key := range keys {
		v := cond[key]
		var expr Expr
		switch e := v.(type) {
		case Expr:
			expr = e
		default:
			expr = Eq(e)
		}
		stmts[i] = &stmt{column: key, expr: expr}
	}
	return &clauseEx{stmts: stmts}
}
