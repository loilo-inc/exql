package main

import (
	"log"

	"github.com/loilo-inc/exql/v3"
	"github.com/loilo-inc/exql/v3/model"
)

func MapJoinedRowsOuterJoin(db exql.DB) {
	query := `
	SELECT * FROM users
	LEFT JOIN group_users ON group_users.user_id = users.id
	LEFT JOIN user_groups ON user_groups.id = group_users.id
	WHERE users.id = ?`
	rows, err := db.DB().Query(query, 1)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer rows.Close()
	splitter := func(i int) string {
		// Each column's separator is `id`
		return "id"
	}
	var users []*model.Users
	var groups []*model.UserGroups
	for rows.Next() {
		var user model.Users
		var groupUser *model.GroupUsers // Use *GroupUsers/*Group for outer join so that it can be nil
		var group *model.UserGroups     // when the values of outer joined columns are NULL.
		if err := exql.MapJoinedRows(db, splitter, rows, &user, &groupUser, &group); err != nil {
			log.Fatal(err.Error())
			return
		}
		users = append(users, &user)
		groups = append(groups, group) // group = nil when the user does not belong to any group.
	}
	// enumerate users and groups.
}
