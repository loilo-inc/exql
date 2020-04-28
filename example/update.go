package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql"
)

func Update() {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = 1
	_, err := db.Update("users", exql.SET{
		"name": "GoGo",
	}, exql.Where("id = ?", 1))
	if err != nil {
		log.Errorf(err.Error())
	}
}
