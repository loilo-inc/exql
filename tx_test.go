package exql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
	"github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
)

func TestTx_Transaction(t *testing.T) {
	db := testDb()
	t.Run("basic", func(t *testing.T) {
		var user *model.Users
		err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
			user = &model.Users{Name: "go"}
			res, err := tx.Insert(user)
			assert.NoError(t, err)
			lid, err := res.LastInsertId()
			assert.NoError(t, err)
			assert.Equal(t, lid, user.Id)
			return nil
		})
		assert.NoError(t, err)
		var dest model.Users
		err = db.Find(query.Q(`select * from users where id = ?`, user.Id), &dest)
		assert.NoError(t, err)
		assert.Equal(t, user.Id, dest.Id)
	})
	t.Run("rollback", func(t *testing.T) {
		var user *model.Users
		err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
			user = &model.Users{Name: "go"}
			res, err := tx.Insert(user)
			assert.NoError(t, err)
			lid, err := res.LastInsertId()
			assert.NoError(t, err)
			assert.Equal(t, lid, user.Id)
			return fmt.Errorf("err")
		})
		assert.EqualError(t, err, "err")
		var dest model.Users
		rows, err := db.DB().Query(`select * from users where id = ?`, user.Id)
		assert.NoError(t, err)
		err = exql.MapRow(rows, &dest)
		assert.ErrorIs(t, err, exql.ErrRecordNotFound)
	})
	t.Run("should rollback if panic happened during transaction", func(t *testing.T) {
		var user *model.Users
		err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
			user = &model.Users{}
			_, err := tx.Insert(user)
			assert.NoError(t, err)
			panic("panic")
		})
		assert.EqualError(t, err, "recovered: panic")
		var dest model.Users
		err = db.Find(query.Q(`select * from users where id = ?`, user.Id), &dest)
		assert.Equal(t, exql.ErrRecordNotFound, err)
	})
}
func TestTx_Map(t *testing.T) {
	db := testDb()
	user := &model.Users{Name: "go"}
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
		if err := exql.MapRow(rows, &dest); err != nil {
			return err
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, user.Id, dest.Id)
}

func TestTx_MapMany(t *testing.T) {
	user := &model.Users{Name: "go"}
	db := testDb()
	var dest []*model.Users
	defer func() {
		db.DB().Exec(`delete from users where id = ?`, user.Id)
	}()
	err := exql.Transaction(db.DB(), context.Background(), nil, func(tx exql.Tx) error {
		if _, err := tx.Insert(user); err != nil {
			return err
		}
		return tx.FindMany(
			query.Q(`select * from users where id = ?`, user.Id),
			&dest,
		)
	})
	assert.NoError(t, err)
	assert.Equal(t, user.Id, dest[0].Id)
}
