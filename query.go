package exql

import (
	"fmt"
	"regexp"
	"strings"
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

type Conditions interface {
	Add(t *Condition)
	IsSafe() bool
	String(prefix *string) (string, error)
	Args() []interface{}
}

type conditions struct {
	conditions []*Condition
}

type Condition struct {
	Text string
	Arg  interface{}
}

func NewConditions(conds []*Condition) Conditions {
	return &conditions{conditions: conds}
}

func (c *conditions) Add(t *Condition) {
	c.conditions = append(c.conditions, t)
}

func (c *conditions) IsSafe() bool {
	return len(c.conditions) > 0
}

func (c *conditions) String(prefix *string) (string, error) {
	if !c.IsSafe() {
		return "", fmt.Errorf("no conditions")
	}

	var sb strings.Builder
	for i, cond := range c.conditions {
		if i > 0 {
			sb.WriteString(" and ")
		}
		if prefix != nil {
			sb.WriteString(fmt.Sprintf("%s.", *prefix))
		}
		sb.WriteString(cond.Text)
	}
	return sb.String(), nil
}

func (c *conditions) Args() []interface{} {
	var values []interface{}
	for _, cond := range c.conditions {
		if cond.Arg != nil {
			values = append(values, cond.Arg)
		}
	}
	return values
}

func (c *conditions) Where(prefix *string) (Clause, error) {
	str, err := c.String(prefix)
	if err != nil {
		return nil, err
	}
	return Where(str, c.Args()...), nil
}
