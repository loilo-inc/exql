package exql

import (
	"context"
	"fmt"
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
	"testing"
)

func TestTx_Transaction(t *testing.T) {
	db := testDb()
	t.Run("basic", func(t *testing.T) {
		var user *model.Users
		err := transaction(db.DB(), context.Background(), nil, func(tx Tx) error {
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
		err := transaction(db.DB(), context.Background(), nil, func(tx Tx) error {
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
		assert.Error(t, err, ErrRecordNotFound.Error())
	})
}
