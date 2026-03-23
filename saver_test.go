package exql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/loilo-inc/exql/v3/mocks/mock_iface"
	"github.com/loilo-inc/exql/v3/mocks/mock_query"
	"github.com/loilo-inc/exql/v3/model"
	"github.com/loilo-inc/exql/v3/model/testmodel"
	q "github.com/loilo-inc/exql/v3/query"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSaver_Insert(t *testing.T) {
	d := testDb()
	s := NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		user := &model.Users{
			Name: "go", Age: 10,
		}
		result, err := s.Insert(user)
		assert.NoError(t, err)
		assert.False(t, user.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, user.Id)
		}()
		r, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, user.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.NoError(t, err)
		var actual model.Users
		err = MapRow(rows, &actual)
		assert.NoError(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, user.Name, actual.Name)
		assert.Equal(t, user.Age, actual.Age)
	})
	t.Run("should error if modelPtr is invalid", func(t *testing.T) {
		res, err := s.Insert(nil)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
	t.Run("should error if db.Exec() failed", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `users`").WithArgs(int64(0), "").WillReturnError(fmt.Errorf("err"))
		s := NewSaver(db)
		user := &model.Users{}
		_, err := s.Insert(user)
		assert.EqualError(t, err, "err")
	})
	t.Run("should error if result.LastInsertId() failed", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `users`").WithArgs(int64(0), "").WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("err")))
		s := NewSaver(db)
		user := &model.Users{}
		_, err := s.Insert(user)
		assert.EqualError(t, err, "err")
	})
	t.Run("should error if auto-increment field type is unsupported", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `samplePrimaryUint64`").WillReturnResult(sqlmock.NewResult(11, 1))
		s := NewSaver(db)
		user := &testmodel.PrimaryUint64{}
		_, err := s.Insert(user)
		assert.ErrorContains(t, err, "auto_increment field must be int64")
	})
	t.Run("should not assign lid in case of not auto_increment", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `sampleNoAutoIncrementKey`").WillReturnResult(sqlmock.NewResult(11, 1))
		s := NewSaver(db)
		user := &testmodel.NoAutoIncrementKey{
			Id: 1,
		}
		_, err := s.Insert(user)
		assert.NoError(t, err)
		assert.Equal(t, 1, user.Id)
	})
}

func TestSaver_InsertContext(t *testing.T) {
	d := testDb()
	s := NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		user := &model.Users{
			Name: "go", Age: 10,
		}
		result, err := s.InsertContext(context.Background(), user)
		assert.NoError(t, err)
		assert.False(t, user.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, user.Id)
		}()
		r, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, user.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.NoError(t, err)
		var actual model.Users
		err = MapRow(rows, &actual)
		assert.NoError(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, user.Name, actual.Name)
		assert.Equal(t, user.Age, actual.Age)
	})
	t.Run("inserting to composite primary key table", func(t *testing.T) {
		history := &model.UserLoginHistories{
			UserId:    1,
			CreatedAt: time.Now(),
		}
		result, err := s.InsertContext(context.Background(), history)
		assert.NoError(t, err)
		assert.False(t, history.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM user_login_histries WHERE id = ?`, history.Id)
		}()
		r, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, history.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM user_login_histories WHERE id = ?`, lid)
		assert.NoError(t, err)
		var actual model.UserLoginHistories
		err = MapRow(rows, &actual)
		assert.NoError(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, history.UserId, actual.UserId)
		assert.Equal(t, history.CreatedAt.Round(time.Second), actual.CreatedAt.Round(time.Second))
	})
	t.Run("inserting to composite primary key table", func(t *testing.T) {
		history := &model.UserLoginHistories{
			UserId:    1,
			CreatedAt: time.Now(),
		}
		result, err := s.InsertContext(context.Background(), history)
		assert.NoError(t, err)
		assert.False(t, history.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM user_login_histries WHERE id = ?`, history.Id)
		}()
		r, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, history.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM user_login_histories WHERE id = ?`, lid)
		assert.NoError(t, err)
		var actual model.UserLoginHistories
		err = MapRow(rows, &actual)
		assert.NoError(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, history.UserId, actual.UserId)
		assert.Equal(t, history.CreatedAt.Round(time.Second), actual.CreatedAt.Round(time.Second))
	})
	t.Run("inserting to no auto_increment key table", func(t *testing.T) {
		user := &model.Users{
			Name: "go", Age: 10,
		}
		result, err := s.InsertContext(context.Background(), user)
		assert.NoError(t, err)
		assert.False(t, user.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, user.Id)
		}()
		r, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, user.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.NoError(t, err)
		var actual model.Users
		err = MapRow(rows, &actual)
		assert.NoError(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, user.Name, actual.Name)
		assert.Equal(t, user.Age, actual.Age)
	})
}

func TestSaver_Update(t *testing.T) {
	d := testDb()
	s := NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		result, err := d.DB().Exec(
			"INSERT INTO `users` (`age`,`name`) VALUES (?, ?)",
			int64(10), "go")
		assert.NoError(t, err)
		lid, err := result.LastInsertId()
		assert.NoError(t, err)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, lid)
		}()
		result, err = s.Update("users", map[string]any{
			"name": "lang",
			"age":  int64(20),
		}, Where(`id = ?`, lid))
		assert.NoError(t, err)
		ra, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), ra)
		var actual model.Users
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.NoError(t, err)
		err = MapRow(rows, &actual)
		assert.NoError(t, err)
		assert.Equal(t, "lang", actual.Name)
		assert.Equal(t, int64(20), actual.Age)
	})
	t.Run("should error if tableName is empty", func(t *testing.T) {
		q, err := s.Update("", nil, nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty table name for update query")
	})
	t.Run("should error if where clause is nil", func(t *testing.T) {
		q, err := s.Update("users", make(map[string]any), nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "nil condition for update query")
	})
	t.Run("should error if map is empty", func(t *testing.T) {
		q, err := s.Update("users", make(map[string]any), Where("id = 1"))
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty values for set clause")
	})
	t.Run("should error if where clause is empty", func(t *testing.T) {
		q, err := s.Update("users", map[string]any{"first_name": "go"}, Where(""))
		assert.Nil(t, q)
		assert.EqualError(t, err, "DANGER: empty query")
	})
}

func TestSaver_UpdateModel(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		s := NewSaver(db)
		name := "lang"
		mock.ExpectExec(
			"UPDATE `users` SET `name` = \\? WHERE id = \\?",
		).WithArgs(name, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		result, err := s.UpdateModel(&model.UpdateUsers{
			Name: &name,
		}, Where(`id = ?`, 1))
		if err != nil {
			t.Fatal(err)
		}
		lid, _ := result.LastInsertId()
		row, _ := result.RowsAffected()
		assert.Equal(t, int64(1), row)
		assert.Equal(t, int64(1), lid)
	})
}

func TestSaver_UpdateModelContext(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		s := NewSaver(db)
		name := "name"
		mock.ExpectExec(
			"UPDATE `users` SET `name` = \\? WHERE id = \\?",
		).WithArgs(name, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		result, err := s.UpdateModelContext(context.Background(), &model.UpdateUsers{
			Name: &name,
		}, Where(`id = ?`, 1))
		if err != nil {
			t.Fatal(err)
		}
		lid, _ := result.LastInsertId()
		row, _ := result.RowsAffected()
		assert.Equal(t, int64(1), row)
		assert.Equal(t, int64(1), lid)
	})
	t.Run("should error if model invalid", func(t *testing.T) {
		db, _, _ := sqlmock.New()
		s := NewSaver(db)
		_, err := s.UpdateModelContext(context.Background(), nil, Where("id = ?", 1))
		assert.EqualError(t, err, "model is nil")
	})
}

func TestSaver_UpdateContext(t *testing.T) {
	d := testDb()
	s := NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		result, err := d.DB().Exec(
			"INSERT INTO `users` (`age`,`name`) VALUES (?, ?)",
			int64(10), "last")
		assert.NoError(t, err)
		lid, err := result.LastInsertId()
		assert.NoError(t, err)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, lid)
		}()
		result, err = s.UpdateContext(context.Background(), "users", map[string]any{
			"age":  int64(20),
			"name": "lang",
		}, Where(`id = ?`, lid))
		assert.NoError(t, err)
		ra, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), ra)
		var actual model.Users
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.NoError(t, err)
		err = MapRow(rows, &actual)
		assert.NoError(t, err)
		assert.Equal(t, "lang", actual.Name)
		assert.Equal(t, int64(20), actual.Age)
	})
}

func TestSaver_Delete(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("DELETE FROM `table` WHERE id = ?").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s := NewSaver(db)
		_, err := s.Delete("table", Where("id = ?", 1))
		assert.NoError(t, err)
	})
	t.Run("should error if clause returened an error", func(t *testing.T) {
		s := NewSaver(nil)
		res, err := s.Delete("table", Where(""))
		assert.EqualError(t, err, "DANGER: empty query")
		assert.Nil(t, res)
	})
	t.Run("should error if table name is empty", func(t *testing.T) {
		s := NewSaver(nil)
		res, err := s.Delete("", Where(""))
		assert.EqualError(t, err, "empty table name for delete query")
		assert.Nil(t, res)
	})
	t.Run("should error if condition is nil ", func(t *testing.T) {
		s := NewSaver(nil)
		res, err := s.Delete("table", nil)
		assert.EqualError(t, err, "nil condition for delete query")
		assert.Nil(t, res)
	})
}

func TestSaver_QueryExtra(t *testing.T) {
	query := q.NewBuilder().Query("SELECT * FROM table WHERE id = ?", 1).Build()
	stmt := "SELECT * FROM table WHERE id = ?"
	args := []any{1}
	aErr := fmt.Errorf("err")
	ctx := context.TODO()
	setup := func(t *testing.T) (*mock_iface.MockExecutor, Saver) {
		ctrl := gomock.NewController(t)
		ex := mock_iface.NewMockExecutor(ctrl)
		s := NewSaver(ex)
		return ex, s
	}
	setupQueryErr := func(t *testing.T) (*mock_query.MockQuery, Saver) {
		ctrl := gomock.NewController(t)
		query := mock_query.NewMockQuery(ctrl)
		query.EXPECT().Query().Return("", nil, aErr)
		s := NewSaver(nil)
		return query, s
	}

	t.Run("Exec", func(t *testing.T) {
		ex, s := setup(t)
		ex.EXPECT().Exec(stmt, args...).Return(nil, nil)
		res, err := s.Exec(query)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("Exec/Error", func(t *testing.T) {
		query, s := setupQueryErr(t)
		res, err := s.Exec(query)
		assert.Nil(t, res)
		assert.Equal(t, aErr, err)
	})
	t.Run("ExecContext", func(t *testing.T) {
		ex, s := setup(t)
		ex.EXPECT().ExecContext(ctx, stmt, args...).Return(nil, nil)
		res, err := s.ExecContext(ctx, query)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("ExecContext/Error", func(t *testing.T) {
		query, s := setupQueryErr(t)
		res, err := s.ExecContext(ctx, query)
		assert.Nil(t, res)
		assert.Equal(t, aErr, err)
	})
	t.Run("Query", func(t *testing.T) {
		ex, s := setup(t)
		ex.EXPECT().Query(stmt, args...).Return(nil, nil)
		res, err := s.Query(query)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("Query/Error", func(t *testing.T) {
		query, s := setupQueryErr(t)
		res, err := s.Query(query)
		assert.Nil(t, res)
		assert.Equal(t, aErr, err)
	})
	t.Run("QueryContext", func(t *testing.T) {
		ex, s := setup(t)
		ex.EXPECT().QueryContext(ctx, stmt, args...).Return(nil, nil)
		res, err := s.QueryContext(ctx, query)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("QueryContext/Error", func(t *testing.T) {
		query, s := setupQueryErr(t)
		res, err := s.QueryContext(ctx, query)
		assert.Nil(t, res)
		assert.Equal(t, aErr, err)
	})
	t.Run("QueryRow", func(t *testing.T) {
		ex, s := setup(t)
		ex.EXPECT().QueryRow(stmt, args...).Return(nil)
		res, err := s.QueryRow(query)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("QueryRow/Error", func(t *testing.T) {
		query, s := setupQueryErr(t)
		res, err := s.QueryRow(query)
		assert.Nil(t, res)
		assert.Equal(t, aErr, err)
	})
	t.Run("QueryRowContext", func(t *testing.T) {
		ex, s := setup(t)
		ex.EXPECT().QueryRowContext(ctx, stmt, args...).Return(nil)
		res, err := s.QueryRowContext(ctx, query)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("QueryRowContext/Error", func(t *testing.T) {
		query, s := setupQueryErr(t)
		res, err := s.QueryRowContext(ctx, query)
		assert.Nil(t, res)
		assert.Equal(t, aErr, err)
	})
}
