package exql

import (
	"context"
	"fmt"
	"io"
)

type logger struct {
	w     io.Writer
	onErr func(error)
}

func (l *logger) Hook(ctx context.Context, query string, args ...any) {
	if _, err := fmt.Fprintf(
		l.w, "%s \t %+v\n", query, args,
	); err != nil && l.onErr != nil {
		l.onErr(err)
	}
}

func NewLogger(w io.Writer, onError func(err error)) Hook {
	return &logger{w: w, onErr: onError}
}
