package exql

import (
	"fmt"
	"testing"

	"github.com/loilo-inc/exql/v3/mocks/mock_query"
	"github.com/loilo-inc/exql/v3/model"
	"github.com/loilo-inc/exql/v3/query"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_genericFinder(t *testing.T) {
	db := testDb()
	ctrl := gomock.NewController(t)

	user1 := model.Users{Name: "user1"}
	user2 := model.Users{Name: "user2"}
	_, err := db.Insert(&user1)
	assert.NoError(t, err)
	_, err = db.Insert(&user2)
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Delete(
			model.UsersTableName,
			query.Cond("id in (?,?)", user1.Id, user2.Id),
		)
	})

	f := newGenericFinder[model.Users](db.DB(), noCacheReflector)

	t.Run("Find", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			dest, err := f.Find(query.Q(`select * from users where id = ?`, user1.Id))
			assert.NoError(t, err)
			if assert.NotNil(t, dest) {
				assert.Equal(t, user1.Id, dest.Id)
				assert.Equal(t, user1.Name, dest.Name)
			}
		})

		t.Run("should error if query is invalid", func(t *testing.T) {
			q := mock_query.NewMockQuery(ctrl)
			q.EXPECT().Query().Return("", nil, fmt.Errorf("err"))
			dest, err := f.Find(q)
			assert.Nil(t, dest)
			assert.EqualError(t, err, "err")
		})

		t.Run("should error if query failed", func(t *testing.T) {
			dest, err := f.Find(query.Q(`select`))
			assert.Nil(t, dest)
			assert.Error(t, err)
		})

		t.Run("should error if mapping failed", func(t *testing.T) {
			dest, err := f.Find(query.Q(`select * from users where id = -1`))
			assert.Nil(t, dest)
			assert.ErrorIs(t, err, ErrRecordNotFound{})
		})
	})

	t.Run("FindMany", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			dest, err := f.FindMany(query.Q(`select * from users where id in (?,?)`, user1.Id, user2.Id))
			assert.NoError(t, err)
			if assert.Len(t, dest, 2) {
				assert.ElementsMatch(t, []int64{user1.Id, user2.Id}, []int64{dest[0].Id, dest[1].Id})
			}
		})

		t.Run("should error if query is invalid", func(t *testing.T) {
			q := mock_query.NewMockQuery(ctrl)
			q.EXPECT().Query().Return("", nil, fmt.Errorf("err"))
			dest, err := f.FindMany(q)
			assert.Nil(t, dest)
			assert.EqualError(t, err, "err")
		})

		t.Run("should error if query failed", func(t *testing.T) {
			dest, err := f.FindMany(query.Q(`select`))
			assert.Nil(t, dest)
			assert.Error(t, err)
		})

		t.Run("should error if mapping failed", func(t *testing.T) {
			dest, err := f.FindMany(query.Q(`select * from users where id = -1`))
			assert.Nil(t, dest)
			assert.ErrorIs(t, err, ErrRecordNotFound{})
		})
	})
}
