package exql

import "github.com/loilo-inc/exql/query"

type Clause = query.Stmt

func Where(q string, args ...any) Clause {
	return query.New(q, args...)
}

func WhereEx(cond map[string]any) Clause {
	return query.QueryEx(cond)
}

func WhereAnd(list ...Clause) Clause {
	return query.QueryAnd(list...)
}
