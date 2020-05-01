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
