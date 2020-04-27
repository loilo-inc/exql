package exql

import (
	"database/sql"
	"github.com/apex/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapper_MapRows(t *testing.T) {
	mapper := &mapper{}
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3326)/exql?charset=utf8mb4&parseTime=True&loc=Local")
	assert.Nil(t, err)
	defer db.Close()

	t.Run("struct", func(t *testing.T) {
		rows, err := db.Query(`SELECT * FROM users LIMIT 1`)
		assert.Nil(t, err)
		defer rows.Close()
		var dest model.Users
		err = mapper.MapRows(rows, &dest)
		assert.Nil(t, err)
		assert.Equal(t, dest.FirstName.String, "go")
		assert.Equal(t, dest.LastName.String, "lang")
	})
	t.Run("slice", func(t *testing.T) {
		rows, err := db.Query(`SELECT * FROM users LIMIT 1`)
		assert.Nil(t, err)
		defer rows.Close()
		var dest []*model.Users
		err = mapper.MapRows(rows, &dest)
		assert.Nil(t, err)
		assert.Equal(t, dest[0].FirstName.String, "go")
		assert.Equal(t, dest[0].LastName.String, "lang")
	})
}

func Test2(t *testing.T) {
	co := func(dest *[]int) {
		for i := 0; i < 10; i++ {
			*dest = append(*dest, i)
		}
	}
	var d []int
	co(&d)
	log.Infof("%+v", d)
}
