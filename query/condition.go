//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package query

type Condition interface {
	Condition() (string, []any, error)
}

type cond struct {
	query string
	args  []interface{}
}

func (w *cond) Condition() (string, []any, error) {
	if err := assertEmptyQuery(w.query); err != nil {
		return "", nil, err
	}
	return w.query, w.args, nil
}

type kvCond struct {
	stmts KeyIterator
}

func (c *kvCond) Condition() (string, []any, error) {
	var arr []Condition
	for i := 0; i < c.stmts.Size(); i++ {
		column, v := c.stmts.Get(i)
		var pred Predicate
		switch e := v.(type) {
		case Predicate:
			pred = e
		default:
			pred = Eq(e)
		}
		if query, args, err := pred.Predicate(column); err != nil {
			return "", nil, err
		} else {
			arr = append(arr, &cond{query: query, args: args})
		}
	}
	return concatConds(kAnd, arr...).Condition()
}

func Where(q string, args ...interface{}) Condition {
	return &cond{
		query: q,
		args:  args,
	}
}

func WhereEx(cond map[string]any) Condition {
	e := NewKeyIterator(cond)
	return &kvCond{stmts: e}
}

type multiCond struct {
	op   string
	list []Condition
}

func (m *multiCond) Condition() (string, []any, error) {
	var list []string
	var values []any
	for _, v := range m.list {
		q, args, err := v.Condition()
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

func WhereAnd(list ...Condition) Condition {
	return concatConds(kAnd, list...)
}
func WhereOr(list ...Condition) Condition {
	return concatConds(kOr, list...)
}

func concatConds(op string, list ...Condition) Condition {
	return &multiCond{op: op, list: list}
}
