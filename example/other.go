package main

import "github.com/loilo-inc/exql/v2"

// db.DB() returns *sql.DB
func OtherQuery(db exql.DB) {
	db.DB().Exec("SELECT * FROM users LIMIT 10")
}
