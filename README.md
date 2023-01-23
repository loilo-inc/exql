exql
---
[![codecov](https://codecov.io/gh/loilo-inc/exql/branch/master/graph/badge.svg?token=aGixN2xIMP)](https://codecov.io/gh/loilo-inc/exql)

Safe, strict and clear ORM for Go

## Introduction

exql is a simple ORM library for MySQL, written in Go. It is designed to work at the minimum for real software development. It has a few, limited but enough convenient functionalities of SQL database.
We adopted the data mapper model, not the active record. Records in the database are mapped into structs simply. Each model has no state and also no methods to modify itself and sync database records. You need to write bare SQL code for every operation you need except for a few cases.

exql is designed by focusing on safety and clearness in SQL usage. In other words, we never generate any SQL statements that are potentially dangerous or have ambiguous side effects across tables and the database.

It does:

- make insert/update query from model structs.
- map rows returned from the database into structs.
- map joined table into one or more structs.
- provide a safe syntax for the transaction.
- provide a framework to build dynamic SQL statements safely.
- generate model codes automatically from the database.

It DOESN'T

- make delete/update statements across the table.
- make unexpectedly slow select queries that don't use correct indices.
- modify any database settings, schemas and indices.

## Table of contents

- [exql](#exql)
- [Introduction](#introduction)
- [Table of contents](#table-of-contents)
- [Usage](#usage)
  - [Open database connection](#open-database-connection)
  - [Generate models from the database](#generate-models-from-the-database)
- [Execute queries](#execute-queries)
  - [Insert](#insert)
  - [Update](#update)
  - [Transaction](#transaction)
- [Map rows into structs](#map-rows-into-structs)
  - [Map rows](#map-rows)
  - [For joined table](#for-joined-table)
  - [For outer-joined table](#for-outer-joined-table)
- [Using query builder](#using-query-builder)
- [License](#license)

## Usage

### Open database connection

```go
package main

import (
	"time"

	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
)

func OpenDB() exql.DB {
	db, err := exql.Open(&exql.OpenOptions{
		// MySQL url for sql.Open()
		Url: "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local",
		// Max retry count for database connection failure
		MaxRetryCount: 3,
		RetryInterval: 10 * time.Second,
	})
	if err != nil {
		log.Fatalf("open error: %s", err)
		return nil
	}
	return db
}

```

### Generate models from the database

```go
package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
)

func GenerateModels() {
	gen := exql.NewGenerator(db.DB())
	err := gen.Generate(&exql.GenerateOptions{
		// Directory path for result. Default is `model`
		OutDir: "dist",
		// Package name for models. Default is `model`
		Package: "dist",
		// Exclude table names for generation. Default is []
		Exclude: []string{
			"internal",
		},
	})
	if err != nil {
		log.Fatalf(err.Error())
	}
}

```

## Execute queries

### Insert

```go
package main

import (
	"github.com/apex/log"
)

func main() {
	// Create user model
	// Primary key (id) is not needed to set. It will be ignored on building insert query.
	user := User{
		Name: "Go",
	}
	// You must pass model as a pointer.
	if result, err := db.Insert(&user); err != nil {
		log.Error(err.Error())
	} else {
		insertedId, _ := result.LastInsertId()
		// Inserted id is inserted into primary key field after insertion, if field is int64/uint64
		if insertedId != user.Id {
			log.Fatalf("impossible")
		}
	}
}

```

### Update

```go
package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
)

func Update() {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = ?
	// [1]
	_, err := db.Update("users", map[string]any{
		"name": "GoGo",
	}, exql.Where("id = ?", 1))
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

```

### Transaction

```go
package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
	"github.com/volatiletech/null"
)

func Transaction() {
	timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := db.TransactionWithContext(timeout, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	}, func(tx exql.Tx) error {
		user := model.Users{
			FirstName: null.String{},
			LastName:  null.String{},
		}
		_, err := tx.Insert(&user)
		return err
	})
	if err != nil {
		// Transaction has been rolled back
	} else {
		// Transaction has been committed
	}
}

```

## Map rows into structs

### Map rows

```go
package main

import "github.com/apex/log"

func Map() {
	// select query
	rows, err := db.DB().Query(`SELECT * FROM users WHERE id = ?`, 1)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		// Destination model struct
		var user User
		// Passing destination to Map(). The second argument must be a pointer of the model.
		if err := db.Map(rows, &user); err != nil {
			log.Error(err.Error())
		}
		log.Infof("%d", user.Id) // -> 1
	}
}

func MapMany() {
	rows, err := db.DB().Query(`SELECT * FROM users LIMIT ?`, 5)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		// Destination slice of models.
		// NOTE: It must be the slice of pointers of models.
		var users []*User
		// Passing destination to MapMany().
		// Second argument must be a pointer.
		if err := db.MapMany(rows, &users); err != nil {
			log.Error(err.Error())
		}
		log.Infof("%d", len(users)) // -> 5
	}
}

```

### For joined table

```go
package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
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

```

### For outer-joined table

```go
package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2"
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
		var user User
		var schoolUser *SchoolUsers // Use *SchoolUsers/*School for outer join so that it can be nil
		var school *School          // when the values of outer joined columns are NULL.
		if err := serialMapper.Map(rows, &user, &schoolUser, &school); err != nil {
			log.Error(err.Error())
			return
		}
		users = append(users, &user)
		schools = append(schools, school) // school = nil when the user does not belong to any school.
	}
	// enumerate users and schools.
}

```

## Using query builder

```go
package main

import (
	"github.com/loilo-inc/exql/v2/query"
)

func Query() {
	query.New(
		`SELECT * FROM users WHERE id IN (:?) AND age = ?`,
		query.V(1, 2, 3), 20,
	)
	// SELECT * FROM users WHERE id IN (?,?,?) AND age = ?
	// [1,2,3,20]
}

func QueryBulider() {
	qb := query.NewBuilder()
	qb.Sprintf("SELECT * FROM %s", "users")
	qb.Query("WHERE id IN (:?) AND age >= ?", query.V(1, 2), 20)
	// SELECT * FROM users WHERE id IN (?,?) AND age >= ?
	// [1,2,20]
	_, _ = db.Query(qb.Build())
}

func CondBulider() {
	cond := query.Cond("id = ?", 1)
	cond.And("age >= ?", 20)
	cond.And("name in (:?)", query.V("go", "lang"))
	q := query.New("SELECT * FROM users WHERE :?", cond)
	// SELECT * FROM users WHERE id = ? and age >= ? and name in (?,?)
	// [1, 20, go, lang]
	_, _ = db.Query(q)
}

```

## License

MIT License / Copyright (c) 2020-Present LoiLo inc.

