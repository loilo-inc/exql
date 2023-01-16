package exfmt_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2/exfmt"
	"github.com/stretchr/testify/assert"
)

func TestFormatter_Noramlize(t *testing.T) {
	var f exfmt.Formatter
	cases := [][]string{
		{"  select\t`users`.`id` from users where name =  'go\"'  \n\n ",
			"SELECT `users`.`id` FROM users WHERE name = 'go\"'",
		},
		{
			"where  (   id in\n ( ?, ? ) )",
			"WHERE ( id IN ( ? , ? ) )",
		},
	}
	for _, v := range cases {
		q, err := f.Normalize(v[0])
		assert.NoError(t, err)
		assert.Equal(t, v[1], q)
	}
}
