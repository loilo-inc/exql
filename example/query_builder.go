package main

import (
	"github.com/loilo-inc/exql/v2/query"
)

func Query() {
	q := query.New(
		`SELECT * FROM users WHERE id IN (:?) AND age = ?`,
		query.V(1, 2, 3), 20,
	)
	// SELECT * FROM users WHERE id IN (?,?,?) AND age = ?
	// [1,2,3,20]
	db.Query(q)
}

func QueryBulider() {
	qb := query.NewBuilder()
	qb.Sprintf("SELECT * FROM %s", "users")
	qb.Query("WHERE id IN (:?) AND age >= ?", query.V(1, 2), 20)
	// SELECT * FROM users WHERE id IN (?,?) AND age >= ?
	// [1,2,20]
	db.Query(qb.Build())
}

func CondBulider() {
	cond := query.Cond("id = ?", 1)
	cond.And("age >= ?", 20)
	cond.And("name in (:?)", query.V("go", "lang"))
	q := query.New("SELECT * FROM users WHERE :?", cond)
	// SELECT * FROM users WHERE id = ? and age >= ? and name in (?,?)
	// [1, 20, go, lang]
	db.Query(q)
}
