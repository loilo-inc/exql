package exql

import (
	"github.com/apex/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
	"testing"
)

func TestMapper_MapRows(t *testing.T) {
	mapper := &mapper{}
	db, err := Open(&OpenOptions{
		Url: "root:@tcp(127.0.0.1:3326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	assert.Nil(t, err)
	user1 := &model.Users{
		FirstName: null.StringFrom("user1"),
		LastName:  null.StringFrom("name"),
	}
	user2 := &model.Users{
		FirstName: null.StringFrom("user2"),
		LastName:  null.StringFrom("name"),
	}
	_, err = db.Insert(user1)
	_, err = db.Insert(user2)
	assert.Nil(t, err)
	defer func() {
		db.DB().Exec(`DELETE FROM users WHERE id = ?`, user1.Id)
		db.DB().Exec(`DELETE FROM users WHERE id = ?`, user2.Id)
		db.Close()
	}()

	t.Run("struct", func(t *testing.T) {
		rows, err := db.DB().Query(`SELECT * FROM users LIMIT 1`)
		assert.Nil(t, err)
		defer rows.Close()
		var dest model.Users
		err = mapper.MapRows(rows, &dest)
		assert.Nil(t, err)
		assert.Equal(t, dest.FirstName.String, "user1")
		assert.Equal(t, dest.LastName.String, "name")
	})
	t.Run("slice", func(t *testing.T) {
		rows, err := db.DB().Query(`SELECT * FROM users LIMIT 2`)
		assert.Nil(t, err)
		defer rows.Close()
		var dest []*model.Users
		err = mapper.MapRows(rows, &dest)
		assert.Nil(t, err)
		assert.Equal(t, dest[0].FirstName.String, "user1")
		assert.Equal(t, dest[0].LastName.String, "name")
		assert.Equal(t, dest[1].FirstName.String, "user2")
		assert.Equal(t, dest[1].LastName.String, "name")
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
