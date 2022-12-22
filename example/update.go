package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
)

func Update() {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = ?
	// [1]
	_, err := db.Update("users", map[string]any{
		"name": "GoGo",
	}, exql.Where("id = ?", 1))
	if err != nil {
		log.Errorf(err.Error())
	}
}

func Delete() {
	// DELETE FROM `users` WHERE id = ?
	// [1]
	_, err := db.Delete("users", exql.Where("id = ?", 1))
	if err != nil {
		log.Errorf(err.Error())
	}
}
