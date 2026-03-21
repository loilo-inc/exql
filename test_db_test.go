package exql_test

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v3"
	"github.com/loilo-inc/exql/v3/test"
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

func noCacheSaver(ex exql.Executor) exql.Saver {
	return exql.NewSaver(ex, exql.NoCacheReflector())
}

func noCacheFinder(ex exql.Executor) exql.Finder {
	return exql.NewFinder(ex, exql.NoCacheReflector())
}
