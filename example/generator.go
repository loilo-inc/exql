package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v2"
)

func GenerateModels() {
	db, _ := sql.Open("mysql", "url-for-db")
	gen := exql.NewGenerator(db)
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
