package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// for stubbing in tests
var timeNow = time.Now

// Create makes an empty pair of up/down migration files in dir, named
// with the current timestamp as the version, and returns their paths.
// dir is created if it does not exist. It fails if a file with the
// same name already exists.
func Create(dir string, name string) ([]string, error) {
	if name == "" || filepath.Base(name) != name {
		return nil, fmt.Errorf("invalid migration name: %q", name)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	version := timeNow().UTC().Format("20060102150405")
	var paths []string
	for _, direction := range []string{"up", "down"} {
		path := filepath.Join(dir, fmt.Sprintf("%s_%s.%s.sql", version, name, direction))
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
}
