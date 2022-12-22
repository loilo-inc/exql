package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql"
)

type School struct {
	Id   int64  `exql:"column:id;primary;not null;auto_increment"`
	Name string `exql:"column:name;not null"`
}
type SchoolUsers struct {
	Id       int64 `exql:"column:id;primary;not null;auto_increment"`
	UserId   int64 `exql:"column:user_id;not null"`
	SchoolId int64 `exql:"column:school_id;not null"`
}

/*
school has many users
users has many schools
*/
func MapSerial() {
	query := `
	SELECT * FROM users
	JOIN school_users ON school_users.user_id = users.id
	JOIN schools ON schools.id = school_users.id
	WHERE schools.id = ?`
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
	var users []*User
	for rows.Next() {
		var user User
		var schoolUser SchoolUsers
		var school School
		// Create serial mapper. It will split joined columns by logical tables.
		// In this case, joined table and destination mappings are:
		// |   users   |       school_users       |   school  |
		// + --------- + ------------------------ + --------- +
		// | id | name | id | user_id | school_id | id | name |
		// + --------- + ------------------------ + --------- +
		// |   &user   |       &schoolUser        |  &school  |
		// + --------- + ------------------------ + --------- +
		if err := serialMapper.Map(rows, &user, &schoolUser, &school); err != nil {
			log.Error(err.Error())
			return
		}
		users = append(users, &user)
	}
	// enumerate users...
}
