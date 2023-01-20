package extool_test

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/loilo-inc/exql/v2/extool"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		var w bytes.Buffer
		l := extool.NewLogger(&w, func(err error) {
			t.Error(err)
		})
		args := []driver.NamedValue{
			{Value: 1},
			{Value: "str"},
			{Value: time.Time{}},
			{Value: sql.NullString{Valid: false}},
		}
		l.HookQuery(context.Background(), "query", args)
		res := w.String()
		assert.Equal(t, "query\n", res)
	})
}
