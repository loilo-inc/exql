// Package migrate provides a minimal database migration tool
// compatible with golang-migrate. Migration state is stored in the
// schema_migrations table (a single row holding the latest version and
// a dirty flag), so databases previously managed by golang-migrate can
// be taken over as-is.
//
// Migration files follow the golang-migrate naming convention:
//
//	<version>_<name>.up.sql
//	<version>_<name>.down.sql
//
// The core logic (Load, Migrator, Create) is independent of any CLI.
// Cli is a thin wrapper for building a migration command on top of
// them, and cmd/migrate is a ready-made command usable via go tool.
package migrate

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
)

// Migration is a pair of up/down SQL statements for a single version.
// A statement that is empty or contains only whitespace is applied as
// a no-op, same as golang-migrate.
type Migration struct {
	// Version is the numeric prefix of the migration file name.
	Version int64
	// Name is the middle part of the migration file name.
	Name string
	// UpStmt is the content of <version>_<name>.up.sql.
	UpStmt string
	// DownStmt is the content of <version>_<name>.down.sql.
	DownStmt string
	// HasDown reports whether the down migration is defined.
	// Down fails on a migration with HasDown = false.
	HasDown bool
}

var migrationFileName = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

// Load reads migration files from the root of fsys and returns them
// sorted by version in ascending order. Files not matching the
// migration naming convention are ignored. It returns an error if
// a version is duplicated or its up migration is missing.
//
// fsys is typically an embed.FS narrowed by fs.Sub, or os.DirFS:
//
//	//go:embed migrations/*.sql
//	var migrationFS embed.FS
//	sub, _ := fs.Sub(migrationFS, "migrations")
//	migrations, err := migrate.Load(sub)
func Load(fsys fs.FS) ([]*Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	byVersion := map[int64]*Migration{}
	hasUp := map[int64]bool{}
	for _, e := range entries {
		match := migrationFileName.FindStringSubmatch(e.Name())
		if match == nil {
			continue
		}
		version, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %s: %w", e.Name(), err)
		}
		stmt, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}
		m, ok := byVersion[version]
		if !ok {
			m = &Migration{Version: version, Name: match[2]}
			byVersion[version] = m
		} else if m.Name != match[2] {
			return nil, fmt.Errorf("duplicate migration version %d: %s and %s", version, m.Name, match[2])
		}
		if match[3] == "up" {
			if hasUp[version] {
				return nil, fmt.Errorf("duplicate up migration for version %d", version)
			}
			hasUp[version] = true
			m.UpStmt = string(stmt)
		} else {
			if m.HasDown {
				return nil, fmt.Errorf("duplicate down migration for version %d", version)
			}
			m.HasDown = true
			m.DownStmt = string(stmt)
		}
	}

	var migrations []*Migration
	for _, m := range byVersion {
		if !hasUp[m.Version] {
			return nil, fmt.Errorf("missing up migration for %d_%s", m.Version, m.Name)
		}
		migrations = append(migrations, m)
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })

	return migrations, nil
}
