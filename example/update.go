package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

// Using designated update struct
func UpdateModel(db exql.DB) {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = ?
	// [1]
	_, err := db.UpdateModel(&model.UpdateUsers{
		Name: exql.Ptr("GoGo"),
	}, exql.Where("id = ?", 1),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// With table name and key-value pairs
func Update(db exql.DB) {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = ?
	// [1]
	_, err := db.Update("users", map[string]any{
		"name": "GoGo",
	}, exql.Where("id = ?", 1))
	if err != nil {
		log.Fatal(err)
	}
}
