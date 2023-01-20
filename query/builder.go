package query

import (
	"fmt"
)

// Builder is dynamic SQL query builder.
type Builder struct {
	qs []Query
}

// Sprintf is short-hand for fmt.Sprintf.
//
// Example:
//
//	b.Sprintf("%s", "go")
//
// is same as:
//
//	b.Query(fmt.Sprintf("%s", "go"))
func (b *Builder) Sprintf(str string, args ...any) *Builder {
	return b.Query(fmt.Sprintf(str, args...))
}

// Query appends given query component and arguments.
//
// Example:
//
//	b.Query(":?", query.V(1,2))
//
// is same as:
//
//	b.Add(query.Q(":?", query.V(1,2)))
func (b *Builder) Query(str string, args ...any) *Builder {
	b.qs = append(b.qs, Q(str, args...))
	return b
}

// Add appends given Queries components.
func (b *Builder) Add(q ...Query) *Builder {
	b.qs = append(b.qs, q...)
	return b
}

// Build constructs final SQL statement, joining by single space(" ").
func (b *Builder) Build() Query {
	return b.Join(" ")
}

// Clone makes a shallow copy of builder.
func (b *Builder) Clone() *Builder {
	return NewBuilder(b.qs...)
}

// Join joins accumulative query components by given separator.
func (b *Builder) Join(sep string) Query {
	c := &chain{joiner: sep}
	c.append(b.qs...)
	return c
}

func NewBuilder(base ...Query) *Builder {
	return &Builder{qs: base}
}
