package exql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const dbUrl = "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local"

func testDb() DB {
	db, err := Open(&OpenOptions{
		Url: dbUrl,
	})
	if err != nil {
		panic(err)
	}
	return db
}

func testSqlDB() *sql.DB {
	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		panic(err)
	}
	return db
}
