package exql

import "database/sql"
import _ "github.com/go-sql-driver/mysql"

func testDb() DB {
	db, err := Open(&OpenOptions{
		Url: "root:@tcp(127.0.0.1:3326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	if err != nil {
		panic(err)
	}
	return db
}

func testSqlDB() *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3326)/exql?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	return db
}
