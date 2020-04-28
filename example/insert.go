package main

import (
	"github.com/apex/log"
)

func main() {
	// Create user model
	// Primary key (id) is not needed to set. It will be ignored on building insert query.
	user := User{
		Name: "Go",
	}
	// You must pass model as a pointer.
	if result, err := db.Insert(&user); err != nil {
		log.Error(err.Error())
	} else {
		insertedId, _ := result.LastInsertId()
		// Inserted id is inserted into primary key field after insertion, if field is int64/uint64
		if insertedId != user.Id {
			log.Fatalf("impossible")
		}
	}
}
