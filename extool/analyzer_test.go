package extool_test

import (
	"context"
	"encoding/json"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/extool"
	"github.com/stretchr/testify/assert"
)

func testDb() exql.DB {
	db, err := exql.Open(&exql.OpenOptions{
		Url: "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	if err != nil {
		panic(err)
	}
	return db
}

func assertJsonEq(t *testing.T, exp any, act any) {
	a, err := json.Marshal(exp)
	assert.NoError(t, err)
	b, err := json.Marshal(act)
	assert.NoError(t, err)
	assert.JSONEq(t, string(a), string(b))
}

func TestAnalyzer(t *testing.T) {
	db := testDb()
	ana := extool.NewAnalyzer(db.DB())
	res, err := ana.Explain(context.Background(), "select * from users")
	assert.NoError(t, err)
	assertJsonEq(t, []*extool.Explain{
		{Id: 1, SelectType: "SIMPLE", Table: "users",
			Type: "ALL", KeyLen: 0, Rows: 0, Filtered: 100},
	}, res)
}
