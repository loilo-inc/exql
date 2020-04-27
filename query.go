package exql

import (
	"fmt"
)

type WhereQuery interface {
	Query() (string, error)
	Args() []interface{}
}

type whereQuery struct {
	query string
	args  []interface{}
}

func (w *whereQuery) Query() (string, error) {
	if w.query == "" {
		return "", fmt.Errorf("empty where clause")
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
