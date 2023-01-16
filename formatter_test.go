package exql_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2"
	"github.com/stretchr/testify/assert"
)

func TestFormatter_Noramlize(t *testing.T) {
	var f exql.Formatter
	cases := [][]string{
		{"  select\t`users`.`id` from users where name =  'go\"'  \n\n ",
			"select `users`.`id` from users where name = 'go\"'",
		},
		{
			"where  ( id in ( ?, ? ) )",
			"where (id in (?,?))",
		},
	}
	for _, v := range cases {
		assert.Equal(t, v[1], f.Normalize(v[0]))
	}
}
