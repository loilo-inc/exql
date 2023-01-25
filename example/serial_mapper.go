package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

/*
user_groups has many users
users belongs to many groups
*/
func MapSerial(db exql.DB) {
	query := `
	SELECT * FROM users
	JOIN group_users ON group_users.user_id = users.id
	JOIN user_groups ON user_groups.id = group_users.id
	WHERE user_groups.name = ?`
	rows, err := db.DB().Query(query, "goland")
	if err != nil {
		log.Fatal(err)
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
		var groupUsers model.GroupUsers
		var userGroup model.UserGroups
		// Create serial mapper. It will split joined columns by logical tables.
		// In this case, joined table and destination mappings are:
		// |   users   |       group_users        |  user_groups  |
		// + --------- + ------------------------ + ------------- +
		// | id | name | id | user_id |  group_id |  id  |  name  |
		// + --------- + ------------------------ + ------------- +
		// |   &user   |        &groupUsers       |   &userGroup  |
		// + --------- + ------------------------ + ------------- +
		if err := serialMapper.Map(rows, &user, &groupUsers, &userGroup); err != nil {
			log.Fatalf(err.Error())
			return
		}
		users = append(users, &user)
	}
	// enumerate users...
}
