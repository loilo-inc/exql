package exql

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
	"testing"
	"time"
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
	Id int `exql:"column:id;primary"`
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
	m := NewMapper()
	s := NewSaver(d.DB())
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
		s := NewSaver(db)
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
		s := NewSaver(db)
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
		s := NewSaver(db)
		user := &samplePrimaryUint64{}
		_, err := s.Insert(user)
		assert.Nil(t, err)
		assert.Equal(t, uint64(11), user.Id)
	})
}

func TestSaver_InsertContext(t *testing.T) {
	d := testDb()
	m := NewMapper()
	s := NewSaver(d.DB())
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
}

func TestSaver_Update(t *testing.T) {
	d := testDb()
	m := NewMapper()
	s := NewSaver(d.DB())
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
		}, Where(`id = ?`, lid))
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

func TestSaver_UpdateContext(t *testing.T) {
	d := testDb()
	m := NewMapper()
	s := NewSaver(d.DB())
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
		}, Where(`id = ?`, lid))
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

func TestSaver_QueryForInsert(t *testing.T) {
	s := &saver{}
	t.Run("basic", func(t *testing.T) {
		user := model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("name"),
		}
		s, err := s.QueryForInsert(&user)
		assert.Nil(t, err)
		exp := "INSERT INTO `users` (`first_name`, `last_name`) VALUES (?, ?)"
		assert.Equal(t, exp, s.Query)
		assert.ElementsMatch(t, s.Values, []interface{}{
			user.FirstName, user.LastName,
		})
	})
	assertInvalid := func(t *testing.T, m interface{}, e string) {
		s, err := s.QueryForInsert(m)
		assert.Nil(t, s)
		assert.EqualError(t, err, e)
	}
	t.Run("should error if dest is nil", func(t *testing.T) {
		assertInvalid(t, nil, "pointer is nil")
	})
	t.Run("should error if dest is not pointer", func(t *testing.T) {
		user := model.Users{}
		assertInvalid(t, user, "object must be pointer of struct")
	})
	t.Run("should error if dest is not pointer of struct", func(t *testing.T) {
		var users []*model.Users
		assertInvalid(t, users, "object must be pointer of struct")
	})
	t.Run("should error if dest has no exql tags in any field", func(t *testing.T) {
		var tim time.Time
		assertInvalid(t, &tim, "obj doesn't have exql tags in any fields")
	})
	t.Run("should error if dest doesn't implement TableName()", func(t *testing.T) {
		var sam sampleNoTableName
		assertInvalid(t, &sam, "obj doesn't implement TableName() method")
	})
	t.Run("should error if TableName() doesn't return string", func(t *testing.T) {
		var sam sampleBadTableName
		assertInvalid(t, &sam, "wrong implementation of TableName()")
	})
	t.Run("should error if field doesn't have column tag", func(t *testing.T) {
		var sam sampleNoColumnTag
		assertInvalid(t, &sam, "column tag is not set")
	})
	t.Run("should error if field tag is invalid", func(t *testing.T) {
		var sam sampleBadTag
		assertInvalid(t, &sam, "duplicated tag: a")
	})
	t.Run("should error if dest has no primary key tag", func(t *testing.T) {
		var sam sampleNoPrimaryKey
		assertInvalid(t, &sam, "table has no primary key")
	})
	t.Run("should error if dest has no auto_increment tag", func(t *testing.T) {
		var sam sampleNoAutoIncrementKey
		assertInvalid(t, &sam, "table has no auto_increment field")
	})
}

func TestSaver_QueryForUpdate(t *testing.T) {
	s := &saver{}
	t.Run("basic", func(t *testing.T) {
		q, err := s.QueryForUpdate("users", map[string]interface{}{
			"beta":  "b",
			"zeta":  "z",
			"alpha": "a",
			"gamma": "g",
		}, Where(`id = ?`, 1))
		assert.Nil(t, err)
		exp := "UPDATE `users` SET `alpha` = ?, `beta` = ?, `gamma` = ?, `zeta` = ? WHERE id = ?"
		assert.Equal(t, exp, q.Query)
		assert.ElementsMatch(t, []string{"alpha", "beta", "gamma", "zeta"}, q.Fields)
		assert.ElementsMatch(t, []interface{}{"a", "b", "g", "z", 1}, q.Values)
	})
	t.Run("should error if tableName is empty", func(t *testing.T) {
		q, err := s.Update("", nil, nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty table name")
	})
	t.Run("should error if where clause is nil", func(t *testing.T) {
		q, err := s.Update("users", make(map[string]interface{}), nil)
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty field set")
	})
	t.Run("should error if where clause is empty", func(t *testing.T) {
		q, err := s.Update("users", map[string]interface{}{
			"first_name": "go",
		}, Where(""))
		assert.Nil(t, q)
		assert.EqualError(t, err, "DANGER: empty where clause")
	})
	t.Run("should error if clause type is not where", func(t *testing.T) {
		q, err := s.Update("users", map[string]interface{}{
			"first_name": "go",
		}, &clause{
			t: "join",
		})
		assert.Nil(t, q)
		assert.EqualError(t, err, "where is not build by Where()")
	})
}
