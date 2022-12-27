package query_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2/query"
)

func TestQuery(t *testing.T) {
	str := []string{"a", "b"}
	query.Vals(str)
}
