package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

func Update() {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = ?
	// [1]
	_, err := db.UpdateModel(&model.UpdateUsers{
		Name: exql.Ptr("GoGo"),
	}, exql.Where("id = ?", 1),
	)
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
