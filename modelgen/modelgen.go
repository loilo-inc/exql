package modelgen

import (
	"flag"
	"log"
	"strings"

	"github.com/loilo-inc/exql/v2"
)

func Generate(driverName string) {
	opts := exql.GenerateOptions{}
	flag.StringVar(&opts.OutDir, "outDir", "model", "Directory path for results")
	flag.StringVar(&opts.Package, "package", "model", "Go package name for result")
	exclude := flag.String("exclude", "", "Comma separated table names to be excluded from results")
	dsn := flag.String("dsn", "", "[Required] DSN (data source name) for the database")
	flag.Parse()
	if *dsn == "" {
		log.Fatalf("missing -dsn option")
	}
	opts.Exclude = strings.Split(*exclude, ",")
	db, err := exql.Open(&exql.OpenOptions{
		DriverName: driverName,
		Url:        *dsn,
	})
	if err != nil {
		log.Fatal(err)
	}
	gen := exql.NewGenerator(db.DB())
	if err := gen.Generate(&opts); err != nil {
		log.Fatal(err)
	}
}
