package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
	"github.com/loilo-inc/exql/v2/query"
)

func Find(db exql.DB) {
	// Destination model struct
	var user model.Users
	// Pass as a pointer
	err := db.Find(query.Q(`SELECT * FROM users WHERE id = ?`, 1), &user)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d", user.Id) // -> 1
}

func FindMany(db exql.DB) {
	// Destination slice of models.
	// NOTE: It must be the slice of pointers of models.
	var users []*model.Users
	// Passing destination to MapMany().
	// Second argument must be a pointer.
	err := db.FindMany(query.Q(`SELECT * FROM users LIMIT ?`, 5), &users)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d", len(users)) // -> 5
}
