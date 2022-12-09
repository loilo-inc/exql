package query

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Stmt() (string, []any, error)
}

type stmt struct {
	query string
	args  []interface{}
}

func (w *stmt) Stmt() (string, []any, error) {
	if err := GuardDangerousQuery(w.query); err != nil {
		return "", nil, err
	}
	return w.query, w.args, nil
}

type stmtEx struct {
	stmts KeyIterator
}

func (c *stmtEx) Stmt() (string, []any, error) {
	var arr []string
	var values []any
	for i := 0; i < c.stmts.Size(); i++ {
		column, v := c.stmts.Get(i)
		var expr Expr
		switch e := v.(type) {
		case Expr:
			expr = e
		default:
			expr = Eq(e)
		}
		if expr, args, err := expr.Expr(column); err != nil {
			return "", nil, err
		} else {
			arr = append(arr, expr)
			values = append(values, args...)
		}
	}
	query := strings.Join(arr, " AND ")
	if err := GuardDangerousQuery(query); err != nil {
		return "", nil, err
	}
	return query, values, nil
}

func NewStmt(q string, args ...interface{}) Stmt {
	return &stmt{
		query: q,
		args:  args,
	}
}

func NewStmtEx(cond map[string]any) Stmt {
	e := NewKeyIterator(cond)
	return &stmtEx{stmts: e}
}

type multiStmt struct {
	clauses []Stmt
}

func (w *multiStmt) Stmt() (string, []any, error) {
	var list []string
	var values []any
	for _, v := range w.clauses {
		q, args, err := v.Stmt()
		if err != nil {
			return "", nil, err
		}
		list = append(list, fmt.Sprintf("(%s)", q))
		values = append(values, args)
	}
	return strings.Join(list, " AND "), values, nil
}

func StmtAnd(list ...Stmt) Stmt {
	return &multiStmt{clauses: list}
}
