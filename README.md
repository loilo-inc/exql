# exql
[![codecov](https://codecov.io/gh/loilo-inc/exql/branch/master/graph/badge.svg?token=aGixN2xIMP)](https://codecov.io/gh/loilo-inc/exql)

Safe, Strict and Clear ORM for Go

## Usage

### Open

```go
package main

import (
	"github.com/apex/log"
	"github.com/loilo-inc/exql"
	"time"
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

### Generate Models

```go
package main

import (
	"github.com/loilo-inc/exql"
	"log"
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
	"github.com/loilo-inc/exql"
)

func Update() {
	// UPDATE `users` SET `name` = `GoGo` WHERE `id` = 1
	_, err := db.Update("users", exql.SET{
		"name": "GoGo",
	}, exql.Where("id = ?", 1))
	if err != nil {
		log.Errorf(err.Error())
	}
}

```

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
		// Passing destination to Map(). Second argument must be a pointer of model struct.
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
		// Destination model structs.
		// NOTE: It must be slice of pointer of model structure
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

### Map joined rows

```go
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

```

#### In case of outer join

```go
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

### Transaction

```go
package main

import (
	"context"
	"database/sql"
	"github.com/loilo-inc/exql"
	"github.com/loilo-inc/exql/model"
	"github.com/volatiletech/null"
	"time"
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
 
