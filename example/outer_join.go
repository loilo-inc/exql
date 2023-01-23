package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

func MapSerialOuterJoin() {
	query := `
	SELECT * FROM users
	LEFT JOIN group_users ON group_users.user_id = users.id
	LEFT JOIN groups ON groups.id = group_users.id
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
	var users []*model.Users
	var groups []*model.Groups
	for rows.Next() {
		var user model.Users
		var groupUser *model.GroupUsers // Use *GroupUsers/*Group for outer join so that it can be nil
		var group *model.Groups         // when the values of outer joined columns are NULL.
		if err := serialMapper.Map(rows, &user, &groupUser, &group); err != nil {
			log.Error(err.Error())
			return
		}
		users = append(users, &user)
		groups = append(groups, group) // group = nil when the user does not belong to any group.
	}
	// enumerate users and groups.
}
