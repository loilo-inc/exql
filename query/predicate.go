//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

type Predicate interface {
	Predicate() (string, []any, error)
}

type predicate struct {
	query string
	args  []interface{}
}

func (w *predicate) Predicate() (string, []any, error) {
	if err := assertEmptyQuery(w.query); err != nil {
		return "", nil, err
	}
	return w.query, w.args, nil
}

type kvPredicate struct {
	stmts KeyIterator
}

func (c *kvPredicate) Predicate() (string, []any, error) {
	var arr []Predicate
	for i := 0; i < c.stmts.Size(); i++ {
		column, v := c.stmts.Get(i)
		var expr Expression
		switch e := v.(type) {
		case Expression:
			expr = e
		default:
			expr = Eq(e)
		}
		if expr, args, err := expr.Expression(column); err != nil {
			return "", nil, err
		} else {
			arr = append(arr, &predicate{query: expr, args: args})
		}
	}
	return concatPredicates(kAnd, arr...).Predicate()
}

func RawPredicate(q string, args ...interface{}) Predicate {
	return &predicate{
		query: q,
		args:  args,
	}
}

func KeyValuePredicate(cond map[string]any) Predicate {
	e := NewKeyIterator(cond)
	return &kvPredicate{stmts: e}
}

type multiPredicate struct {
	op   string
	list []Predicate
}

func (m *multiPredicate) Predicate() (string, []any, error) {
	var list []string
	var values []any
	for _, v := range m.list {
		q, args, err := v.Predicate()
		if err != nil {
			return "", nil, err
		}
		list = append(list, q)
		values = append(values, args...)
	}
	if ret, err := concatQueries(m.op, list); err != nil {
		return "", nil, err
	} else {
		return ret, values, nil
	}
}

func PredicateAnd(list ...Predicate) Predicate {
	return concatPredicates(kAnd, list...)
}
func PredicateOr(list ...Predicate) Predicate {
	return concatPredicates(kOr, list...)
}

func concatPredicates(op string, list ...Predicate) Predicate {
	return &multiPredicate{op: op, list: list}
}
