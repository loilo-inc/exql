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
		// Package name for models. Default is `package`
		Package: "",
		// Exclude table names for generation. Default is []
		Exclude: []string{
			"internal",
		},
	})
	if err != nil {
		log.Fatalf(err.Error())
	}
}
