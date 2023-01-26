package extool

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/query"
)

type Analyzer struct {
	ex      exql.Executor
	f       exql.Finder
	results []*ExplainResult
}

type ExplainResult struct {
	Query  string
	Result []*Explain
	Error  error
}

func (l *Analyzer) Hook(ctx context.Context, q string, args ...any) {
	results, err := l.Explain(ctx, q, args...)
	l.results = append(l.results, &ExplainResult{
		Query: q, Result: results, Error: err,
	})
}

func (l *Analyzer) Results() []*ExplainResult {
	return l.results
}

func (l *Analyzer) Assert(t *testing.T) {
	for _, v := range l.results {
		for _, e := range v.Result {
			switch e.Type {
			case "ALL":
				t.Logf("warn: table scan found for table `%s`", e.Table)
			case "index":
				t.Logf("warn: index found for table `%s`", e.Table)
			}
		}
	}
}

type Explain struct {
	Id           int            `exql:"column:id"`
	SelectType   string         `exql:"column:select_type"`
	Table        string         `exql:"column:table"`
	Partitons    sql.NullString `exql:"column:partitions"`
	Type         string         `exql:"column:type"`
	PollibseKeys sql.NullString `exql:"possible_keys"`
	Key          sql.NullString `exql:"key"`
	KeyLen       int            `exql:"key_len"`
	Ref          sql.NullString `exql:"column:ref"`
	Rows         int            `exql:"column:row"`
	Filtered     float64        `exql:"column:filtered"`
	Extra        sql.NullString `exql:"column:Extra"`
}

func (l *Analyzer) Explain(ctx context.Context, q string, args ...any) ([]*Explain, error) {
	var dest []*Explain
	if err := l.f.FindManyContext(ctx, query.Q(fmt.Sprintf("EXPLAIN %s", q), args...), &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func NewAnalyzer(ex exql.Executor) *Analyzer {
	return &Analyzer{ex: ex, f: exql.NewFinder(ex)}
}
