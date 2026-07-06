package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// Cli is a thin command-line frontend for the migration logic.
// Embed your migration files and call Run from your main function:
//
//	//go:embed migrations/*.sql
//	var migrationFS embed.FS
//
//	func main() {
//		db, err := sql.Open("mysql", dsn)
//		...
//		sub, _ := fs.Sub(migrationFS, "migrations")
//		cli := &migrate.Cli{DB: db, FS: sub, Dir: "migrations"}
//		if err := cli.Run(context.Background(), os.Args[1:]...); err != nil {
//			log.Fatal(err)
//		}
//	}
type Cli struct {
	// Database to migrate. Not required by the create command.
	DB *sql.DB
	// File system containing migration files at its root.
	// Not required by the create command.
	FS fs.FS
	// @default "migrations"
	// Directory where the create command puts new migration files.
	Dir string
	// @default os.Stdout
	// Destination for command output.
	Out io.Writer
	// Options passed to New. Can be nil.
	Options *Options
}

const usage = `usage: <command>

commands:
  up             apply all pending migrations
  down           revert all applied migrations (destructive)
  drop           drop all tables in the database (destructive)
  version        print the current migration version
  create <name>  create a new pair of migration files
`

// Run executes a subcommand: up, down, drop, version, or create.
// The down and drop commands are destructive; guard them by environment
// in your main function if necessary.
func (c *Cli) Run(ctx context.Context, args ...string) error {
	out := c.Out
	if out == nil {
		out = os.Stdout
	}
	if len(args) == 0 {
		return fmt.Errorf("expects a subcommand\n%s", usage)
	}

	if args[0] == "create" {
		if len(args) != 2 {
			return fmt.Errorf("usage: create <name>")
		}
		dir := c.Dir
		if dir == "" {
			dir = "migrations"
		}
		paths, err := Create(dir, args[1])
		if err != nil {
			return err
		}
		for _, p := range paths {
			fmt.Fprintln(out, p)
		}
		return nil
	}

	migrations, err := Load(c.FS)
	if err != nil {
		return err
	}
	opts := c.Options
	if opts == nil {
		opts = &Options{
			Log: func(msg string) { fmt.Fprintln(out, msg) },
		}
	}
	m := New(c.DB, migrations, opts)

	switch args[0] {
	case "up":
		return m.Up(ctx)
	case "down":
		return m.Down(ctx)
	case "drop":
		return m.Drop(ctx)
	case "version":
		version, dirty, err := m.Version(ctx)
		if err != nil {
			return err
		}
		if dirty {
			fmt.Fprintf(out, "%d (dirty)\n", version)
		} else {
			fmt.Fprintf(out, "%d\n", version)
		}
		return nil
	default:
		return fmt.Errorf("unknown command: %s\n%s", args[0], usage)
	}
}
