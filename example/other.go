package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
)

// To execute other kind of queries, unwrap sql.DB.
func OtherQuery(db exql.DB) {
	// db.DB() returns *sql.DB
	row := db.DB().QueryRow("SELECT COUNT(*) FROM users")
	var count int
	row.Scan(&count)
	log.Printf("%d", count)
}
