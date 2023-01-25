package exql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/mocks/mock_query"
	"github.com/loilo-inc/exql/v2/model"
	"github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
)

func TestFinder_Find(t *testing.T) {
	t.Run("call FindContext", func(t *testing.T) {

	})
}

func TestFinder_FindContext(t *testing.T) {
	db := testDb()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	user1 := model.Users{Name: "user1"}
	user2 := model.Users{Name: "user2"}
	db.Insert(&user1)
	db.Insert(&user2)
	t.Cleanup(func() {
		db.Delete(
			model.UsersTableName,
			query.Cond("id in (:?)", query.V(user1.Id, user2.Id)),
		)
	})
	f := exql.NewFinder(db.DB())
	t.Run("basic", func(t *testing.T) {
		var dest model.Users
		err := f.FindContext(ctx, query.Q(`select * from users where id = ?`, user1.Id), &dest)
		assert.NoError(t, err)
		assert.Equal(t, user1.Name, dest.Name)
	})
	t.Run("should error if query is invalid", func(t *testing.T) {
		q := mock_query.NewMockQuery(ctrl)
		q.EXPECT().Query().Return("", nil, fmt.Errorf("err"))
		err := f.FindContext(ctx, q, nil)
		assert.EqualError(t, err, "err")
	})
	t.Run("should error if query failed", func(t *testing.T) {
		err := f.FindContext(ctx, query.Q(`select`), nil)
		assert.Error(t, err)
	})
	t.Run("should error if mapping failed", func(t *testing.T) {
		var dest model.Users
		err := f.FindContext(ctx, query.Q(`select * from users where id = -1`), &dest)
		assert.ErrorIs(t, err, exql.ErrRecordNotFound)
	})
}
