package main

import (
	"github.com/apex/log"
	q "github.com/loilo-inc/exql/query"
)

func Update() {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = ?
	// [1]
	_, err := db.Update("users", map[string]any{
		"name": "GoGo",
	}, q.Where("id = ?", 1))
	if err != nil {
		log.Errorf(err.Error())
	}
}

func Delete() {
	// DELETE FROM `users` WHERE id = ?
	// [1]
	_, err := db.Delete("users", q.Where("id = ?", 1))
	if err != nil {
		log.Errorf(err.Error())
	}
}
