package migrate

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	stubTimeNow(t, time.Date(2026, 7, 6, 12, 34, 56, 0, time.UTC))
	t.Run("creates an empty pair of migration files", func(t *testing.T) {
		dir := t.TempDir()
		paths, err := Create(dir, "create_users")
		assert.NoError(t, err)
		assert.Equal(t, []string{
			filepath.Join(dir, "20260706123456_create_users.up.sql"),
			filepath.Join(dir, "20260706123456_create_users.down.sql"),
		}, paths)
		for _, p := range paths {
			assert.FileExists(t, p)
		}
	})
	t.Run("fails if the file already exists", func(t *testing.T) {
		dir := t.TempDir()
		_, err := Create(dir, "create_users")
		require.NoError(t, err)
		_, err = Create(dir, "create_users")
		assert.ErrorContains(t, err, "file exists")
	})
	t.Run("rejects an invalid name", func(t *testing.T) {
		for _, name := range []string{"", "a/b", "../evil"} {
			_, err := Create(t.TempDir(), name)
			assert.ErrorContains(t, err, "invalid migration name")
		}
	})
}
