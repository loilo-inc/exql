package exql

import (
	"fmt"
	"regexp"
)

type ClauseType string

const (
	ClauseTypeWhere = "where"
)

type Clause interface {
	Query() (string, error)
	Args() []interface{}
	Type() ClauseType
}

type clause struct {
	t     ClauseType
	query string
	args  []interface{}
}

func (w *clause) Type() ClauseType {
	return w.t
}

var emptyPat = regexp.MustCompile("^[\\s|\\t]*$")

func IsSafeWhereClause(s string) bool {
	if emptyPat.MatchString(s) {
		return false
	}
	return true
}

func (w *clause) Query() (string, error) {
	if !IsSafeWhereClause(w.query) {
		return "", fmt.Errorf("DANGER: empty where clause")
	}
	return w.query, nil
}

func (w *clause) Args() []interface{} {
	return w.args
}

func Where(q string, args ...interface{}) Clause {
	return &clause{
		t:     ClauseTypeWhere,
		query: q,
		args:  args,
	}
}
