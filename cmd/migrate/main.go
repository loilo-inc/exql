// Command migrate is a minimal database migration CLI compatible with
// golang-migrate, built on the migrate package. Install it as a Go tool
// and run it with go tool:
//
//	go get -tool github.com/loilo-inc/exql/v3/cmd/migrate
//	go tool migrate -dsn "root:@tcp(127.0.0.1:3306)/app" up
//	go tool migrate create add_users
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/v3/migrate"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flags.SetOutput(out)
	dsn := flags.String("dsn", "", "MySQL DSN. Append multiStatements=true to run migration files containing multiple statements.")
	dir := flags.String("dir", "migrations", "directory containing migration files")
	if err := flags.Parse(args); err != nil {
		return err
	}

	cli := &migrate.Cli{Dir: *dir, Out: out}
	cmd := flags.Args()
	// create only generates files and does not need a database
	if len(cmd) > 0 && cmd[0] != "create" {
		if *dsn == "" {
			return fmt.Errorf("-dsn is required")
		}
		db, err := sql.Open("mysql", *dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		cli.DB = db
		cli.FS = os.DirFS(*dir)
	}

	return cli.Run(ctx, cmd...)
}
