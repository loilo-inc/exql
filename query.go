package exql

import (
	"fmt"
	"regexp"
)

type WhereQuery interface {
	Query() (string, error)
	Args() []interface{}
}

type whereQuery struct {
	query string
	args  []interface{}
}

var emptyPat = regexp.MustCompile("^[\\s|\\t]*$")

func IsSafeWhereClause(s string) bool {
	if emptyPat.MatchString(s) {
		return false
	}
	return true
}

func (w *whereQuery) Query() (string, error) {
	if !IsSafeWhereClause(w.query) {
		return "", fmt.Errorf("DANGER: empty where clause")
	}
	return w.query, nil
}

func (w *whereQuery) Args() []interface{} {
	return w.args
}

func Where(q string, args ...interface{}) WhereQuery {
	return &whereQuery{
		query: q,
		args:  args,
	}
}
