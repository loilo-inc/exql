package query

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Query() (string, error)
	Args() []interface{}
}

type stmt struct {
	query string
	args  []interface{}
}

func (w *stmt) Query() (string, error) {
	return GuardDangerousQuery(w.query)
}

func (w *stmt) Args() []interface{} {
	return w.args
}

type clauseEx struct {
	stmts KeyIterator
}

func (c *clauseEx) Args() []interface{} {
	return c.stmts.Values()
}

func (c *clauseEx) Query() (string, error) {
	var arr []string
	for i := 0; i < c.stmts.Size(); i++ {
		column, v := c.stmts.Get(i)
		var expr Expr
		switch e := v.(type) {
		case Expr:
			expr = e
		default:
			expr = Eq(e)
		}
		if expr, err := expr.Expr(column); err != nil {
			return "", err
		} else {
			arr = append(arr, expr)
		}
	}
	query := strings.Join(arr, " AND ")
	return GuardDangerousQuery(query)
}

func New(q string, args ...interface{}) Stmt {
	return &stmt{
		query: q,
		args:  args,
	}
}

func QueryEx(cond map[string]any) Stmt {
	e := NewKeyIterator(cond)
	return &clauseEx{stmts: e}
}

type whereAnd struct {
	clauses []Stmt
}

func (w *whereAnd) Args() []interface{} {
	var args []any
	for _, v := range w.clauses {
		args = append(args, v.Args()...)
	}
	return args
}

func (w *whereAnd) Query() (string, error) {
	var list []string
	for _, v := range w.clauses {
		q, err := v.Query()
		if err != nil {
			return "", err
		}
		list = append(list, fmt.Sprintf("(%s)", q))
	}
	return strings.Join(list, " AND "), nil
}

func QueryAnd(list ...Stmt) Stmt {
	return &whereAnd{clauses: list}
}
