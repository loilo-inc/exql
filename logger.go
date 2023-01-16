package exql

import (
	"context"
	"fmt"
	"io"

	"github.com/loilo-inc/exql/v2/exfmt"
)

type logger struct {
	w     io.Writer
	f     exfmt.Formatter
	onErr func(error)
}

func (l *logger) Hook(ctx context.Context, query string, args ...any) {
	if q, err := l.f.Normalize(query); err != nil {
		l.onErr(err)
	} else if _, err := fmt.Fprintf(l.w, "%s\n", q); err != nil {
		l.onErr(err)
	}
}

func NewLogger(w io.Writer, onError func(err error)) Hook {
	if onError == nil {
		onError = func(err error) {}
	}
	return &logger{w: w, onErr: onError}
}
