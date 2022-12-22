package exql_test

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v2"
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

func testDbMySQL8() exql.DB {
	db, err := exql.Open(&exql.OpenOptions{
		Url: "root:@tcp(127.0.0.1:13327)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	if err != nil {
		panic(err)
	}
	return db
}

func testSqlDB() *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	return db
}
