package factory_test

import (
	"testing"

	"github.com/loilo-inc/exql"
	"github.com/loilo-inc/exql/factory"
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
)

func testDb() exql.DB {
	db, err := exql.Open(&exql.OpenOptions{
		Url: "root:@tcp(127.0.0.1:13326)/exql?charset=utf8mb4&parseTime=True&loc=Local",
	})
	if err != nil {
		panic(err)
	}
	return db
}
func TestD_Create(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		d := factory.Domain(t, testDb())
		user := &model.Users{}
		assert.NotPanics(t, func() {
			d.Create(user)
		})
		assert.True(t, user.Id != 0)
	})
	t.Run("should panic when passing model that created already", func(t *testing.T) {
		d := factory.Domain(t, testDb())
		user := &model.Users{}
		assert.NotPanics(t, func() {
			d.Create(user)
		})
		assert.True(t, user.Id != 0)
		assert.Panics(t, func() {
			d.Create(user)
		})
	})
}

func TestD_Save(t *testing.T) {
	db := testDb()
	t.Run("basic", func(t *testing.T) {
		d := factory.Domain(t, db)
		user := &model.Users{}
		assert.NotPanics(t, func() {
			d.Create(user)
		})
		user.LastName = null.StringFrom("last")
		user.FirstName = null.StringFrom("first")
		assert.NotPanics(t, func() {
			d.Save(user)
		})
		rows, err := db.DB().Query(`select * from users where id = ?`, user.Id)
		assert.Nil(t, err)
		var dest model.Users
		assert.Nil(t, db.Map(rows, &dest))
		assert.Equal(t, user.LastName.String, dest.LastName.String)
		assert.Equal(t, user.FirstName.String, dest.FirstName.String)
	})
	t.Run("should panic when saving model not created by D", func(t *testing.T) {
		d := factory.Domain(t, db)
		user := &model.Users{}
		assert.Panics(t, func() {
			d.Save(user)
		})
	})
}
