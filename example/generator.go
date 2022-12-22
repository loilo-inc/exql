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
