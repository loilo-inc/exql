package exql

import (
	"context"
	"fmt"
	"io"
)

type logger struct {
	w     io.Writer
	f     Formatter
	onErr func(error)
}

func (l *logger) Hook(ctx context.Context, query string, args ...any) {
	if _, err := fmt.Fprintf(
		l.w, "%s\n", l.f.Normalize(query),
	); err != nil && l.onErr != nil {
		l.onErr(err)
	}
}

func NewLogger(w io.Writer, onError func(err error)) Hook {
	return &logger{w: w, onErr: onError}
}
