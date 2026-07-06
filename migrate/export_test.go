package migrate

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func stubTimeNow(t *testing.T, tm time.Time) {
	orig := timeNow
	timeNow = func() time.Time { return tm }
	t.Cleanup(func() { timeNow = orig })
}

// testDB makes a dedicated database for the test and returns a
// connection to it. The database is dropped when the test finishes.
func testDB(t *testing.T, name string) *sql.DB {
	t.Helper()
	root, err := sql.Open("mysql", "root:@tcp(127.0.0.1:13326)/")
	require.NoError(t, err)
	_, err = root.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", name))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := root.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name))
		require.NoError(t, err)
		require.NoError(t, root.Close())
	})

	db, err := sql.Open("mysql", fmt.Sprintf("root:@tcp(127.0.0.1:13326)/%s?multiStatements=true", name))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}

func tableNames(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(
		"SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' ORDER BY table_name",
	)
	require.NoError(t, err)
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	require.NoError(t, rows.Err())
	return tables
}
