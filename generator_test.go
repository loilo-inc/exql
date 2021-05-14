package exql

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
)

func TestGenerator_Generate(t *testing.T) {
	db := testDb()
	g := NewGenerator(db.DB())
	checkFiles := func(dir string, elements []string) {
		entries, err := os.ReadDir(dir)
		assert.Nil(t, err)
		var files []string
		for _, e := range entries {
			files = append(files, e.Name())
		}
		assert.ElementsMatch(t, files, elements)
	}
	t.Run("bsaic", func(t *testing.T) {
		dir, _ := ioutil.TempDir(os.TempDir(), "dist")
		err := g.Generate(&GenerateOptions{
			OutDir:  dir,
			Package: "dist",
		})
		if err != nil {
			log.Errorf(err.Error())
		}
		assert.Nil(t, err)
		checkFiles(dir, []string{"users.go", "user_groups.go", "group_users.go", "fields.go"})
	})
	t.Run("exclude", func(t *testing.T) {
		dir, _ := ioutil.TempDir(os.TempDir(), "dist")
		err := g.Generate(&GenerateOptions{
			OutDir:  dir,
			Package: "dist",
			Exclude: []string{"fields"},
		})
		if err != nil {
			log.Errorf(err.Error())
		}
		assert.Nil(t, err)
		checkFiles(dir, []string{"users.go", "user_groups.go", "group_users.go"})
	})
}
