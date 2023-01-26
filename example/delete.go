package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
)

func Delete(db exql.DB) {
	// DELETE FROM `users` WHERE id = ?
	// [1]
	_, err := db.Delete("users", exql.Where("id = ?", 1))
	if err != nil {
		log.Fatal(err)
	}
}
