package main

import "github.com/apex/log"

func Map() {
	// select query
	rows, err := db.DB().Query(`SELECT * FROM users WHERE id = ?`, 1)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		// Destination model struct
		var user User
		// Passing destination to Map(). Second argument must be a pointer of model struct.
		if err := db.Map(rows, &user); err != nil {
			log.Error(err.Error())
		}
		log.Infof("%d", user.Id) // -> 1
	}
}

func MapMany() {
	rows, err := db.DB().Query(`SELECT * FROM users LIMIT ?`, 5)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		// Destination model structs.
		// NOTE: It must be slice of pointer of model structure
		var users []*User
		// Passing destination to MapMany().
		// Second argument must be a pointer.
		if err := db.MapMany(rows, &users); err != nil {
			log.Error(err.Error())
		}
		log.Infof("%d", len(users)) // -> 5
	}
}
