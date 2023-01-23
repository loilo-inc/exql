package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

/*
groups has many users
users belongs to many groups
*/
func MapSerial() {
	query := `
	SELECT * FROM users
	JOIN group_users ON group_users.user_id = users.id
	JOIN groups ON groups.id = group_users.id
	WHERE groups.name = ?`
	rows, err := db.DB().Query(query, "goland")
	if err != nil {
		log.Errorf("err")
		return
	}
	defer rows.Close()
	serialMapper := exql.NewSerialMapper(func(i int) string {
		// Each column's separator is `id`
		return "id"
	})
	var users []*model.Users
	for rows.Next() {
		var user model.Users
		var group_users model.GroupUsers
		var group model.Groups
		// Create serial mapper. It will split joined columns by logical tables.
		// In this case, joined table and destination mappings are:
		// |   users   |       group_users        |  groups   |
		// + --------- + ------------------------ + --------- +
		// | id | name | id | user_id |  group_id | id | name |
		// + --------- + ------------------------ + --------- +
		// |   &user   |       &group_users       |  &group   |
		// + --------- + ------------------------ + --------- +
		if err := serialMapper.Map(rows, &user, &group_users, &group); err != nil {
			log.Error(err.Error())
			return
		}
		users = append(users, &user)
	}
	// enumerate users...
}
