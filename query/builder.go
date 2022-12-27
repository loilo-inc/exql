package query

import (
	"fmt"
)

type Builder struct {
	qs []Query
}

func (b *Builder) Sprintf(str string, args ...any) *Builder {
	return b.Query(fmt.Sprintf(str, args...))
}

func (b *Builder) Qprintf(str string, args ...Query) *Builder {
	return b.Add(&fmtQuery{fmt: str, qs: args})
}

func (b *Builder) Query(str string, args ...any) *Builder {
	b.qs = append(b.qs, NewQuery(str, args...))
	return b
}

func (b *Builder) Args(args ...any) *Builder {
	b.qs = append(b.qs, NewQuery("", args...))
	return b
}

func (b *Builder) Add(q ...Query) *Builder {
	b.qs = append(b.qs, q...)
	return b
}

func (b *Builder) Build() Query {
	return b.Join(" ")
}

func (b *Builder) Clone() *Builder {
	return NewBuilder(b.qs...)
}

func (b *Builder) Join(sep string) Query {
	return &chain{op: sep, qs: b.qs}
}

func NewBuilder(base ...Query) *Builder {
	return &Builder{qs: base}
}
