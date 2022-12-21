package main

import q "github.com/loilo-inc/exql/query"

func UseQuery() {
	selectQuery := q.Select{
		From:  "users",
		Where: q.NewCondition("id = ?", 1),
	}
	selectStmt, selectArgs, _ := selectQuery.Query()
	db.DB().Query(
		selectStmt,    // SELECT * FROM `users` WHERE id = ?
		selectArgs..., // [1]

	)
	insertQuery := q.Insert{
		Into: "users",
		Values: map[string]any{
			"age":  10,
			"name": "go",
		},
	}
	insertStmt, insertArgs, _ := insertQuery.Query()
	db.DB().Exec(
		insertStmt,    // INSERT INTO `users` (`age`,`name`) VALUES (?,?)
		insertArgs..., // [10, "go"]
	)

	updateQuery := q.Update{
		Table: "users",
		Set: map[string]any{
			"age":  20,
			"name": "go",
		},
		Where: q.NewCondition("id = ?", 1),
	}
	updateStmt, updateArgs, _ := updateQuery.Query()
	db.DB().Exec(
		updateStmt,    // UPDATE `users` SET `age` = ?,`name` = ? WHERE id = ?
		updateArgs..., // [20,"go",1]
	)

	deleteQuery := q.Delete{
		From:  "users",
		Where: q.NewCondition("id = ?", 1),
	}
	deleteStmt, deleteArgs, _ := deleteQuery.Query()
	db.DB().Exec(
		deleteStmt,    // DELETE FROM `users` WHERE id = ?
		deleteArgs..., // [1]
	)
}
