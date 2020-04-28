package exql

import (
	"github.com/loilo-inc/exql/model"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
	"testing"
	"time"
)

type sample1 struct {
	Id int `exql:"column:id;primary"`
}
type sample2 struct {
	Id int `exql:"column:id;primary"`
}

func (s *sample2) TableName() interface{} {
	return 1
}

type sample3 struct {
	Id int `exql:"column:id"`
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

func TestSaver_QueryForInsert(t *testing.T) {
	s := &saver{}
	t.Run("basic", func(t *testing.T) {
		user := model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("name"),
		}
		s, err := s.QueryForInsert(&user)
		assert.Nil(t, err)
		exp := "INSERT INTO `users` (first_name, last_name) VALUES (?, ?)"
		assert.Equal(t, exp, s.Query)
		assert.ElementsMatch(t, s.Values, []interface{}{
			user.FirstName, user.LastName,
		})
	})
	t.Run("objがポインタでない場合エラーを返す", func(t *testing.T) {
		user := model.Users{}
		s, err := s.QueryForInsert(user)
		assert.Nil(t, s)
		assert.Errorf(t, err, "object must be pointer of struct")
	})
	t.Run("objが構造体のポインタでない場合エラーを返す", func(t *testing.T) {
		var users []*model.Users
		s, err := s.QueryForInsert(&users)
		assert.Nil(t, s)
		assert.Errorf(t, err, "object must be pointer of struct")
	})
	t.Run("exqlタグのない構造体のポインタの場合エラーを返す", func(t *testing.T) {
		var tim time.Time
		s, err := s.QueryForInsert(&tim)
		assert.Nil(t, s)
		assert.Errorf(t, err, "obj doesn't have exql tags in any fields")
	})
	t.Run("TableName()がない構造体の場合エラーを返す", func(t *testing.T) {
		var sam sample1
		s, err := s.QueryForInsert(&sam)
		assert.Nil(t, s)
		assert.Errorf(t, err, "obj doesn't implement TableName() method")
	})
	t.Run("TableName()が文字列を返さない場合エラーを返す", func(t *testing.T) {
		var sam sample2
		s, err := s.QueryForInsert(&sam)
		assert.Nil(t, s)
		assert.Errorf(t, err, "wrong implementation of TableName()")
	})
	t.Run("主キーのないモデルはエラーを返す", func(t *testing.T) {
		var sam sample3
		s, err := s.QueryForInsert(&sam)
		assert.Nil(t, s)
		assert.Errorf(t, err, "table has no primary key")
	})
}

func TestSaver_QueryForUpdate(t *testing.T) {
	s := &saver{}
	t.Run("basic", func(t *testing.T) {
		q, err := s.QueryForUpdate("users", map[string]interface{}{
			"first_name": "go",
			"last_name":  "land",
		}, Where(`id = ?`, 1))
		assert.Nil(t, err)
		exp := "UPDATE `users` SET `first_name` = ?, `last_name` = ? WHERE id = ?"
		assert.Equal(t, exp, q.Query)
		assert.ElementsMatch(t, []string{"first_name", "last_name"}, q.Fields)
		assert.ElementsMatch(t, []interface{}{"go", "land", 1}, q.Values)
	})
	t.Run("tableが空の場合はエラー", func(t *testing.T) {
		q, err := s.Update("", nil, nil)
		assert.Nil(t, q)
		assert.Errorf(t, err, "empty table name")
	})
	t.Run("setが空の場合はエラー", func(t *testing.T) {
		q, err := s.Update("users", make(map[string]interface{}), nil)
		assert.Nil(t, q)
		assert.Errorf(t, err, "empty field set")
	})
	t.Run("Whereが空の場合はエラー", func(t *testing.T) {
		q, err := s.Update("users", map[string]interface{}{
			"first_name": "go",
		}, Where(""))
		assert.Nil(t, q)
		assert.Errorf(t, err, "empty where clause")
	})
	t.Run("return error if clause type is not where", func(t *testing.T) {
		q, err := s.Update("users", map[string]interface{}{
			"first_name": "go",
		}, &clause{
			t: "join",
		})
		assert.Nil(t, q)
		assert.Errorf(t, err, "where is not build by Where()")
	})
}
