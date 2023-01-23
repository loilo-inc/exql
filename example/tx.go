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
