package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql"
)

func MapSerialOuterJoin() {
	query := `
	SELECT * FROM users
	LEFT JOIN school_users ON school_users.user_id = users.id
	LEFT JOIN schools ON schools.id = school_users.id
	WHERE users.id = ?`
	rows, err := db.DB().Query(query, 1)
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
	var schools []*School
	for rows.Next() {
		var user *User              // Use *Model so that mapper can set *Model = nil
		var schoolUser *SchoolUsers // when the separator column's value is NULL
		var school *School          // in the joined columns.
		if err := serialMapper.Map(rows, &user, &schoolUser, &school); err != nil {
			log.Error(err.Error())
			return
		}
		users = append(users, user)
		schools = append(schools, school) // school = nil when the user does not belong to any school.
	}
	// enumerate users and schools.
}
