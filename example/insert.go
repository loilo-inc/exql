package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

func Insert() {
	// Create a user model
	// Primary key (id) is not needed to set.
	// It will be ignored on building the insert query.
	user := model.Users{
		Name: "Go",
	}
	// You must pass the model as a pointer.
	if result, err := db.Insert(&user); err != nil {
		log.Fatal(err.Error())
	} else {
		insertedId, _ := result.LastInsertId()
		// Inserted id is assigned into the auto-increment field after the insertion,
		// if these field is int64/uint64
		if insertedId != user.Id {
			log.Fatal("never happens")
		}
	}
}

func BulkInsert() {
	user1 := model.Users{Name: "Go"}
	user2 := model.Users{Name: "Lang"}
	// INSERT INTO users (name) VALUES (?),(?)
	// ["Go", "Lang"]
	if q, err := exql.QueryForBulkInsert(user1, user2); err != nil {
		log.Fatal(err)
	} else if _, err := db.Exec(q); err != nil {
		log.Fatal(err)
	}
	// NOTE: unlike a single insertion, bulk insertion doesn't obtain auto-incremented values from results.
}
