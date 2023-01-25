package main

import (
	"time"

	"log"

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
