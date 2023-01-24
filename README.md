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
  - [Code Generation](#code-generation)
    - [Generate models from the database](#generate-models-from-the-database)
    - [Auto-Generated code](#auto-generated-code)
  - [Execute queries](#execute-queries)
    - [Insert](#insert)
    - [Update](#update)
  - [Transaction](#transaction)
  - [Map rows into structs](#map-rows-into-structs)
    - [Map rows](#map-rows)
    - [For joined table](#for-joined-table)
    - [For outer-joined table](#for-outer-joined-table)
  - [Use query builder](#use-query-builder)
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

### Code Generation

#### Generate models from the database

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

#### Auto-Generated code

```go
<no value>
```

### Execute queries

#### Insert

```go
package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

func Insert() {
	// Create a user model
	// Primary key (id) is not needed to set.
	// It will be ignored on building the insert query.
	user := model.Users{
		Name: "Go",
	}
	// You must pass the model as a pointer.
	if result, err := db.Insert(&user); err != nil {
		log.Fatal(err.Error())
	} else {
		insertedId, _ := result.LastInsertId()
		// Inserted id is assigned into the auto-increment field after the insertion,
		// if these field is int64/uint64
		if insertedId != user.Id {
			log.Fatal("never happens")
		}
	}
}

func BulkInsert() {
	user1 := model.Users{Name: "Go"}
	user2 := model.Users{Name: "Lang"}
	// INSERT INTO users (name) VALUES (?),(?)
	// ["Go", "Lang"]
	if q, err := exql.QueryForBulkInsert(&user1, &user2); err != nil {
		log.Fatal(err)
	} else if _, err := db.Exec(q); err != nil {
		log.Fatal(err)
	}
	// NOTE: unlike a single insertion, bulk insertion doesn't obtain auto-incremented values from results.
}

```

#### Update

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
)

func Transaction() {
	timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := db.TransactionWithContext(timeout, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	}, func(tx exql.Tx) error {
		user := model.Users{Name: "go"}
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

### Map rows into structs

#### Map rows

```go
package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql/v2/model"
)

func Map() {
	// select query
	rows, err := db.DB().Query(`SELECT * FROM users WHERE id = ?`, 1)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		// Destination model struct
		var user model.Users
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
		var users []*model.Users
		// Passing destination to MapMany().
		// Second argument must be a pointer.
		if err := db.MapMany(rows, &users); err != nil {
			log.Error(err.Error())
		}
		log.Infof("%d", len(users)) // -> 5
	}
}

```

#### For joined table

```go
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
func MapSerial() {
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

```

#### For outer-joined table

```go
package main

import (
	"log"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
)

func MapSerialOuterJoin() {
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
	serialMapper := exql.NewSerialMapper(func(i int) string {
		// Each column's separator is `id`
		return "id"
	})
	var users []*model.Users
	var groups []*model.UserGroups
	for rows.Next() {
		var user model.Users
		var groupUser *model.GroupUsers // Use *GroupUsers/*Group for outer join so that it can be nil
		var group *model.UserGroups     // when the values of outer joined columns are NULL.
		if err := serialMapper.Map(rows, &user, &groupUser, &group); err != nil {
			log.Fatal(err.Error())
			return
		}
		users = append(users, &user)
		groups = append(groups, group) // group = nil when the user does not belong to any group.
	}
	// enumerate users and groups.
}

```

### Use query builder

```go
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

```

## License

MIT License / Copyright (c) 2020-Present LoiLo inc.

