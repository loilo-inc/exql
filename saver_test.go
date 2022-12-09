package exql_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/loilo-inc/exql"
	"github.com/loilo-inc/exql/model"
	"github.com/loilo-inc/exql/query"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
)

type sampleNoTableName struct {
	Id int `exql:"column:id;primary;auto_increment"`
}

type sampleBadTableName struct {
	Id int `exql:"column:id;primary;auto_increment"`
}

func (s *sampleBadTableName) TableName() interface{} {
	return 1
}

type sampleNoPrimaryKey struct {
	Id int `exql:"column:id;auto_increment"`
}

type sampleNoColumnTag struct {
	Id int `exql:"primary;auto_increment"`
}

type sampleBadTag struct {
	Id int `exql:"a;a:1"`
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
		assert.Equal(t, query.ErrDangerousExpr, err)
		assert.Nil(t, res)
	})
}

func TestSaver_QueryForUpdate(t *testing.T) {
	s := exql.NewSaver(nil)
	t.Run("should error if tableName is empty", func(t *testing.T) {
		q, err := s.Update("", nil, nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty table")
	})
	t.Run("should error if where clause is nil", func(t *testing.T) {
		q, err := s.Update("users", make(map[string]interface{}), nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty values")
	})
	t.Run("should error if where clause is empty", func(t *testing.T) {
		q, err := s.Update("users", map[string]interface{}{
			"first_name": "go",
		}, exql.Where(""))
		assert.Nil(t, q)
		assert.EqualError(t, err, "DANGER: empty where clause")
	})
}

type upSampleInvalidTag struct {
	Id *int `exql:"column::"`
}
type upSampleNotPtr struct {
	Id int `exql:"column:id"`
}
type upSample struct {
	Id *int `exql:"column:id"`
}
type upSampleNoFields struct {
}
type upSampleWrongImpl struct {
	Id *int `exql:"column:id"`
}

func (upSampleWrongImpl) ForTableName() int {
	return 1
}

type upSampleNoColumn struct {
	Id *int `exql:"row:id"`
}

func (upSampleNoColumn) ForTableName() string {
	return "table"
}
