package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

func Map(db exql.DB) {
	// select query
	rows, err := db.DB().Query(`SELECT * FROM users WHERE id = ?`, 1)
	if err != nil {
		log.Fatal(err)
	} else {
		// Destination model struct
		var user model.Users
		// Passing destination to Map(). The second argument must be a pointer of the model.
		if err := db.Map(rows, &user); err != nil {
			log.Fatal(err)
		}
		log.Printf("%d", user.Id) // -> 1
	}
}

func MapMany(db exql.DB) {
	rows, err := db.DB().Query(`SELECT * FROM users LIMIT ?`, 5)
	if err != nil {
		log.Fatal(err)
	} else {
		// Destination slice of models.
		// NOTE: It must be the slice of pointers of models.
		var users []*model.Users
		// Passing destination to MapMany().
		// Second argument must be a pointer.
		if err := db.MapMany(rows, &users); err != nil {
			log.Fatal(err)
		}
		log.Printf("%d", len(users)) // -> 5
	}
}
