package exql_test

import (
	"fmt"
	"testing"

	"github.com/loilo-inc/exql/v3"
	"github.com/loilo-inc/exql/v3/mocks/mock_query"
	"github.com/loilo-inc/exql/v3/model"
	"github.com/loilo-inc/exql/v3/query"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFinder(t *testing.T) {
	db := testDb()
	ctrl := gomock.NewController(t)
	user1 := model.Users{Name: "user1"}
	user2 := model.Users{Name: "user2"}
	db.Insert(&user1)
	db.Insert(&user2)
	t.Cleanup(func() {
		db.Delete(
			model.UsersTableName,
			query.Cond("id in (?,?)", user1.Id, user2.Id),
		)
	})
	f := exql.NewFinder(db.DB(), db)
	t.Run("FindContext", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			var dest model.Users
			err := f.Find(query.Q(`select * from users where id = ?`, user1.Id), &dest)
			assert.NoError(t, err)
			assert.Equal(t, user1.Name, dest.Name)
		})
		t.Run("should error if query is invalid", func(t *testing.T) {
			q := mock_query.NewMockQuery(ctrl)
			q.EXPECT().Query().Return("", nil, fmt.Errorf("err"))
			err := f.Find(q, nil)
			assert.EqualError(t, err, "err")
		})
		t.Run("should error if query failed", func(t *testing.T) {
			err := f.Find(query.Q(`select`), nil)
			assert.Error(t, err)
		})
		t.Run("should error if mapping failed", func(t *testing.T) {
			var dest model.Users
			err := f.Find(query.Q(`select * from users where id = -1`), &dest)
			assert.ErrorIs(t, err, exql.ErrRecordNotFound{})
		})
	})
	t.Run("FindManyContext", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			var dest []*model.Users
			err := f.FindMany(query.Q(`select * from users where id in (?,?)`, user1.Id, user2.Id), &dest)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(dest))
			assert.ElementsMatch(t, []int64{user1.Id, user2.Id}, []int64{dest[0].Id, dest[1].Id})
		})
		t.Run("should error if query is invalid", func(t *testing.T) {
			q := mock_query.NewMockQuery(ctrl)
			q.EXPECT().Query().Return("", nil, fmt.Errorf("err"))
			err := f.FindMany(q, nil)
			assert.EqualError(t, err, "err")
		})
		t.Run("should error if query failed", func(t *testing.T) {
			err := f.FindMany(query.Q(`select`), nil)
			assert.Error(t, err)
		})
		t.Run("should error if mapping failed", func(t *testing.T) {
			var dest []*model.Users
			err := f.FindMany(query.Q(`select * from users where id = -1`), &dest)
			assert.ErrorIs(t, err, exql.ErrRecordNotFound{})
		})
	})
}
