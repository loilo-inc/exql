package exql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
)

func TestTx_Transaction(t *testing.T) {
	db := testDb()
	t.Run("basic", func(t *testing.T) {
		var user *model.Users
		err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
			user = &model.Users{
				FirstName: null.StringFrom("go"),
				LastName:  null.StringFrom("land"),
			}
			res, err := tx.Insert(user)
			assert.Nil(t, err)
			lid, err := res.LastInsertId()
			assert.Nil(t, err)
			assert.Equal(t, lid, user.Id)
			return nil
		})
		assert.Nil(t, err)
		var dest model.Users
		rows, err := db.DB().Query(`select * from users where id = ?`, user.Id)
		assert.Nil(t, err)
		assert.Nil(t, db.Map(rows, &dest))
		assert.Equal(t, user.Id, dest.Id)
	})
	t.Run("rollback", func(t *testing.T) {
		var user *model.Users
		err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
			user = &model.Users{
				FirstName: null.StringFrom("go"),
				LastName:  null.StringFrom("land"),
			}
			res, err := tx.Insert(user)
			assert.Nil(t, err)
			lid, err := res.LastInsertId()
			assert.Nil(t, err)
			assert.Equal(t, lid, user.Id)
			return fmt.Errorf("err")
		})
		assert.Error(t, err, "err")
		var dest model.Users
		rows, err := db.DB().Query(`select * from users where id = ?`, user.Id)
		assert.Nil(t, err)
		err = db.Map(rows, &dest)
		assert.Error(t, err, exql.ErrRecordNotFound.Error())
	})
	t.Run("should rollback if panic happened during transaction", func(t *testing.T) {
		var user *model.Users
		err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
			user = &model.Users{
				FirstName: null.String{},
				LastName:  null.String{},
			}
			_, err := tx.Insert(user)
			assert.Nil(t, err)
			panic("panic")
		})
		assert.EqualError(t, err, "recovered: panic")
		rows, err := db.DB().Query(`select * from users where id = ?`, user.Id)
		assert.Nil(t, err)
		var dest model.Users
		assert.Equal(t, db.Map(rows, &dest), exql.ErrRecordNotFound)
	})
}
func TestTx_Map(t *testing.T) {
	db := testDb()
	user := &model.Users{
		FirstName: null.StringFrom("go"),
		LastName:  null.StringFrom("land"),
	}
	defer func() {
		db.DB().Exec(`delete from users where id = ?`, user.Id)
	}()
	var dest model.Users
	err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
		if _, err := tx.Insert(user); err != nil {
			return err
		}
		rows, err := tx.Tx().Query(`select * from users where id = ?`, user.Id)
		if err != nil {
			return err
		}
		if err := tx.Map(rows, &dest); err != nil {
			return err
		}
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, user.Id, dest.Id)
}

func TestTx_MapMany(t *testing.T) {
	user := &model.Users{
		FirstName: null.StringFrom("go"),
		LastName:  null.StringFrom("land"),
	}
	db := testDb()
	var dest []*model.Users
	defer func() {
		db.DB().Exec(`delete from users where id = ?`, user.Id)
	}()
	err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
		if _, err := tx.Insert(user); err != nil {
			return err
		}
		rows, err := tx.Tx().Query(`select * from users where id = ?`, user.Id)
		if err != nil {
			return err
		}
		if err := tx.MapMany(rows, &dest); err != nil {
			return err
		}
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, user.Id, dest[0].Id)
}
