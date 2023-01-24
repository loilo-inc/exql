package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v2"
)

func main() {
	db, _ := exql.Open(&exql.OpenOptions{
		Url: "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	g := exql.NewGenerator(db.DB())
	err := g.Generate(&exql.GenerateOptions{
		OutDir: "model",
	})
	if err != nil {
		log.Fatal(err)
	}
}
