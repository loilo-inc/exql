package exql_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/mocks/mock_exql"
	"github.com/loilo-inc/exql/v2/mocks/mock_query"
	"github.com/loilo-inc/exql/v2/model"
	q "github.com/loilo-inc/exql/v2/query"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
)

type sampleBadTableName struct {
	Id int `exql:"column:id;primary;auto_increment"`
}

func (sampleBadTableName) TableName() string {
	return ""
}

type sampleNoPrimaryKey struct {
	Id int `exql:"column:id;auto_increment"`
}

func (sampleNoPrimaryKey) TableName() string {
	return ""
}

type sampleNoColumnTag struct {
	Id int `exql:"primary;auto_increment"`
}

func (sampleNoColumnTag) TableName() string {
	return ""
}

type sampleBadTag struct {
	Id int `exql:"a;a:1"`
}

func (sampleBadTag) TableName() string {
	return ""
}

type sampleNoAutoIncrementKey struct {
	Id   int    `exql:"column:id;primary"`
	Name string `exql:"column:name"`
}

func (s *sampleNoAutoIncrementKey) TableName() string {
	return "sampleNoAutoIncrementKey"
}

type samplePrimaryUint64 struct {
	Id   uint64 `exql:"column:id;primary;auto_increment"`
	Name string `exql:"column:name"`
}

func (s *samplePrimaryUint64) TableName() string {
	return "samplePrimaryUint64"
}

func TestSaver_Insert(t *testing.T) {
	d := testDb()
	m := exql.NewMapper()
	s := exql.NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		user := &model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("last"),
		}
		result, err := s.Insert(user)
		assert.Nil(t, err)
		assert.False(t, user.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, user.Id)
		}()
		r, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.Nil(t, err)
		assert.Equal(t, user.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.Nil(t, err)
		var actual model.Users
		err = m.Map(rows, &actual)
		assert.Nil(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, user.FirstName.String, actual.FirstName.String)
		assert.Equal(t, user.LastName.String, actual.LastName.String)
	})
	t.Run("should error if modelPtr is invalid", func(t *testing.T) {
		res, err := s.Insert(nil)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
	t.Run("should error if db.Exec() failed", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `users`").WithArgs(null.String{}, null.String{}).WillReturnError(fmt.Errorf("err"))
		s := exql.NewSaver(db)
		user := &model.Users{
			FirstName: null.String{},
			LastName:  null.String{},
		}
		_, err := s.Insert(user)
		assert.EqualError(t, err, "err")
	})
	t.Run("should error if result.LastInsertId() failed", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `users`").WithArgs(null.String{}, null.String{}).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("err")))
		s := exql.NewSaver(db)
		user := &model.Users{
			FirstName: null.String{},
			LastName:  null.String{},
		}
		_, err := s.Insert(user)
		assert.EqualError(t, err, "err")
	})
	t.Run("should assign lid to uint primary key", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `samplePrimaryUint64`").WillReturnResult(sqlmock.NewResult(11, 1))
		s := exql.NewSaver(db)
		user := &samplePrimaryUint64{}
		_, err := s.Insert(user)
		assert.Nil(t, err)
		assert.Equal(t, uint64(11), user.Id)
	})
	t.Run("should not assign lid in case of not auto_increment", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("INSERT INTO `sampleNoAutoIncrementKey`").WillReturnResult(sqlmock.NewResult(11, 1))
		s := exql.NewSaver(db)
		user := &sampleNoAutoIncrementKey{
			Id: 1,
		}
		_, err := s.Insert(user)
		assert.Nil(t, err)
		assert.Equal(t, 1, user.Id)
	})
}

func TestSaver_InsertContext(t *testing.T) {
	d := testDb()
	m := exql.NewMapper()
	s := exql.NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		user := &model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("last"),
		}
		result, err := s.InsertContext(context.Background(), user)
		assert.Nil(t, err)
		assert.False(t, user.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, user.Id)
		}()
		r, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.Nil(t, err)
		assert.Equal(t, user.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.Nil(t, err)
		var actual model.Users
		err = m.Map(rows, &actual)
		assert.Nil(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, user.FirstName.String, actual.FirstName.String)
		assert.Equal(t, user.LastName.String, actual.LastName.String)
	})
	t.Run("inserting to composite primary key table", func(t *testing.T) {
		history := &model.UserLoginHistories{
			UserId:    1,
			CreatedAt: time.Now(),
		}
		result, err := s.InsertContext(context.Background(), history)
		assert.Nil(t, err)
		assert.False(t, history.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM user_login_histries WHERE id = ?`, history.Id)
		}()
		r, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.Nil(t, err)
		assert.Equal(t, history.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM user_login_histories WHERE id = ?`, lid)
		assert.Nil(t, err)
		var actual model.UserLoginHistories
		err = m.Map(rows, &actual)
		assert.Nil(t, err)
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
		assert.Nil(t, err)
		assert.False(t, history.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM user_login_histries WHERE id = ?`, history.Id)
		}()
		r, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.Nil(t, err)
		assert.Equal(t, history.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM user_login_histories WHERE id = ?`, lid)
		assert.Nil(t, err)
		var actual model.UserLoginHistories
		err = m.Map(rows, &actual)
		assert.Nil(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, history.UserId, actual.UserId)
		assert.Equal(t, history.CreatedAt.Round(time.Second), actual.CreatedAt.Round(time.Second))
	})
	t.Run("inserting to no auto_increment key table", func(t *testing.T) {
		user := &model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("last"),
		}
		result, err := s.InsertContext(context.Background(), user)
		assert.Nil(t, err)
		assert.False(t, user.Id == 0)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, user.Id)
		}()
		r, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r)
		lid, err := result.LastInsertId()
		assert.Nil(t, err)
		assert.Equal(t, user.Id, lid)
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.Nil(t, err)
		var actual model.Users
		err = m.Map(rows, &actual)
		assert.Nil(t, err)
		assert.Equal(t, lid, actual.Id)
		assert.Equal(t, user.FirstName.String, actual.FirstName.String)
		assert.Equal(t, user.LastName.String, actual.LastName.String)
	})
}

func TestSaver_Update(t *testing.T) {
	d := testDb()
	m := exql.NewMapper()
	s := exql.NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		result, err := d.DB().Exec(
			"INSERT INTO `users` (`first_name`, `last_name`) VALUES (?, ?)",
			"first", "last")
		assert.Nil(t, err)
		lid, err := result.LastInsertId()
		assert.Nil(t, err)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, lid)
		}()
		result, err = s.Update("users", map[string]interface{}{
			"first_name": "go",
			"last_name":  "lang",
		}, exql.Where(`id = ?`, lid))
		assert.Nil(t, err)
		ra, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), ra)
		var actual model.Users
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.Nil(t, err)
		err = m.Map(rows, &actual)
		assert.Nil(t, err)
		assert.Equal(t, "go", actual.FirstName.String)
		assert.Equal(t, "lang", actual.LastName.String)
	})
	t.Run("should error if tableName is empty", func(t *testing.T) {
		q, err := s.Update("", nil, nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty table name for update query")
	})
	t.Run("should error if where clause is nil", func(t *testing.T) {
		q, err := s.Update("users", make(map[string]interface{}), nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "nil condition for update query")
	})
	t.Run("should error if map is empty", func(t *testing.T) {
		q, err := s.Update("users", make(map[string]interface{}), exql.Where("id = 1"))
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty values for set clause")
	})
	t.Run("should error if where clause is empty", func(t *testing.T) {
		q, err := s.Update("users", map[string]interface{}{"first_name": "go"}, exql.Where(""))
		assert.Nil(t, q)
		assert.EqualError(t, err, "DANGER: empty query")
	})
}

func TestSaver_UpdateModel(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		s := exql.NewSaver(db)
		firstName := null.StringFrom("name")
		mock.ExpectExec(
			"UPDATE `users` SET `first_name` = \\? WHERE id = \\?",
		).WithArgs(firstName, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		result, err := s.UpdateModel(&model.UpdateUsers{
			FirstName: &firstName,
		}, exql.Where(`id = ?`, 1))
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
		s := exql.NewSaver(db)
		firstName := null.StringFrom("name")
		mock.ExpectExec(
			"UPDATE `users` SET `first_name` = \\? WHERE id = \\?",
		).WithArgs(firstName, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		result, err := s.UpdateModelContext(context.Background(), &model.UpdateUsers{
			FirstName: &firstName,
		}, exql.Where(`id = ?`, 1))
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
		s := exql.NewSaver(db)
		_, err := s.UpdateModelContext(context.Background(), nil, exql.Where("id = ?", 1))
		assert.EqualError(t, err, "pointer is nil")
	})
}

func TestSaver_UpdateContext(t *testing.T) {
	d := testDb()
	m := exql.NewMapper()
	s := exql.NewSaver(d.DB())
	t.Run("basic", func(t *testing.T) {
		result, err := d.DB().Exec(
			"INSERT INTO `users` (`first_name`, `last_name`) VALUES (?, ?)",
			"first", "last")
		assert.Nil(t, err)
		lid, err := result.LastInsertId()
		assert.Nil(t, err)
		defer func() {
			d.DB().Exec(`DELETE FROM users WHERE id = ?`, lid)
		}()
		result, err = s.UpdateContext(context.Background(), "users", map[string]interface{}{
			"first_name": "go",
			"last_name":  "lang",
		}, exql.Where(`id = ?`, lid))
		assert.Nil(t, err)
		ra, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), ra)
		var actual model.Users
		rows, err := d.DB().Query(`SELECT * FROM users WHERE id = ?`, lid)
		assert.Nil(t, err)
		err = m.Map(rows, &actual)
		assert.Nil(t, err)
		assert.Equal(t, "go", actual.FirstName.String)
		assert.Equal(t, "lang", actual.LastName.String)
	})
}

func TestSaver_Delete(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		mock.ExpectExec("DELETE FROM `table` WHERE id = ?").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		s := exql.NewSaver(db)
		_, err := s.Delete("table", exql.Where("id = ?", 1))
		assert.NoError(t, err)
	})
	t.Run("should error if clause returened an error", func(t *testing.T) {
		s := exql.NewSaver(nil)
		res, err := s.Delete("table", exql.Where(""))
		assert.EqualError(t, err, "DANGER: empty query")
		assert.Nil(t, res)
	})
	t.Run("should error if table name is empty", func(t *testing.T) {
		s := exql.NewSaver(nil)
		res, err := s.Delete("", exql.Where(""))
		assert.EqualError(t, err, "empty table name for delete query")
		assert.Nil(t, res)
	})
	t.Run("should error if condition is nil ", func(t *testing.T) {
		s := exql.NewSaver(nil)
		res, err := s.Delete("table", nil)
		assert.EqualError(t, err, "nil condition for delete query")
		assert.Nil(t, res)
	})
}

type upSampleInvalidTag struct {
	Id *int `exql:"column::"`
}

func (upSampleInvalidTag) UpdateTableName() string {
	return ""
}

type upSampleNotPtr struct {
	Id int `exql:"column:id"`
}

func (upSampleNotPtr) UpdateTableName() string {
	return ""
}

type upSample struct {
	Id *int `exql:"column:id"`
}

func (upSample) UpdateTableName() string {
	return ""
}

type upSampleNoFields struct {
}

func (upSampleNoFields) UpdateTableName() string {
	return ""
}

type upSampleNoColumn struct {
	Id *int `exql:"row:id"`
}

func (upSampleNoColumn) UpdateTableName() string {
	return "table"
}

func TestSaver_QueryExtra(t *testing.T) {
	query := q.NewBuilder().Query("SELECT * FROM table WHERE id = ?", 1).Build()
	stmt := "SELECT * FROM table WHERE id = ?"
	args := []any{1}
	aErr := fmt.Errorf("err")
	ctx := context.TODO()
	setup := func(t *testing.T) (*mock_exql.MockExecutor, exql.Saver) {
		ctrl := gomock.NewController(t)
		ex := mock_exql.NewMockExecutor(ctrl)
		s := exql.NewSaver(ex)
		return ex, s
	}
	setupQueryErr := func(t *testing.T) (*mock_query.MockQuery, exql.Saver) {
		ctrl := gomock.NewController(t)
		query := mock_query.NewMockQuery(ctrl)
		query.EXPECT().Query().Return("", nil, aErr)
		s := exql.NewSaver(nil)
		return query, s
	}

	t.Run("Exec", func(t *testing.T) {
		ex, s := setup(t)
		ex.EXPECT().Exec(stmt, args...).Return(nil, nil)
		res, err := s.Exec(query)
		assert.Nil(t, res)
		assert.Nil(t, err)
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
		assert.Nil(t, err)
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
		assert.Nil(t, err)
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
		assert.Nil(t, err)
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
		assert.Nil(t, err)
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
		assert.Nil(t, err)
	})
	t.Run("QueryRowContext/Error", func(t *testing.T) {
		query, s := setupQueryErr(t)
		res, err := s.QueryRowContext(ctx, query)
		assert.Nil(t, res)
		assert.Equal(t, aErr, err)
	})
}
