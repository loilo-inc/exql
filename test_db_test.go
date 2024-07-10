package exql_test

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/test"
)

func testDb() exql.DB {
	db, err := exql.Open(&exql.OpenOptions{
		Url: test.DbUrl,
	})
	if err != nil {
		panic(err)
	}
	return db
}

func testSqlDB() *sql.DB {
	db, err := sql.Open("mysql", test.DbUrl)
	if err != nil {
		panic(err)
	}
	return db
}
