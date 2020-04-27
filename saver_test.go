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

func TestBuildInsertQuery(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		user := model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("name"),
		}
		s, err := BuildInsertQuery(&user)
		assert.Nil(t, err)
		exp := "INSERT INTO `users` (first_name, last_name) VALUES (?, ?)"
		assert.Equal(t, exp, s.query)
		assert.ElementsMatch(t, s.values, []interface{}{
			user.FirstName, user.LastName,
		})
	})
	t.Run("objがポインタでない場合エラーを返す", func(t *testing.T) {
		user := model.Users{}
		s, err := BuildInsertQuery(user)
		assert.Nil(t, s)
		assert.Errorf(t, err, "object must be pointer of struct")
	})
	t.Run("objが構造体のポインタでない場合エラーを返す", func(t *testing.T) {
		var users []*model.Users
		s, err := BuildInsertQuery(&users)
		assert.Nil(t, s)
		assert.Errorf(t, err, "object must be pointer of struct")
	})
	t.Run("exqlタグのない構造体のポインタの場合エラーを返す", func(t *testing.T) {
		var tim time.Time
		s, err := BuildInsertQuery(&tim)
		assert.Nil(t, s)
		assert.Errorf(t, err, "obj doesn't have exql tags in any fields")
	})
	t.Run("TableName()がない構造体の場合エラーを返す", func(t *testing.T) {
		var sam sample1
		s, err := BuildInsertQuery(&sam)
		assert.Nil(t, s)
		assert.Errorf(t, err, "obj doesn't implement TableName() method")
	})
	t.Run("TableName()が文字列を返さない場合エラーを返す", func(t *testing.T) {
		var sam sample2
		s, err := BuildInsertQuery(&sam)
		assert.Nil(t, s)
		assert.Errorf(t, err, "wrong implementation of TableName()")
	})
	t.Run("主キーのないモデルはエラーを返す", func(t *testing.T) {
		var sam sample3
		s, err := BuildInsertQuery(&sam)
		assert.Nil(t, s)
		assert.Errorf(t, err, "table has no primary key")
	})
}

func TestBuildUpdateQuery(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		q, err := BuildUpdateQuery("users", map[string]interface{}{
			"first_name": "go",
			"last_name":  "land",
		}, Where(`id = ?`, 1))
		assert.Nil(t, err)
		exp := "UPDATE `users` SET `first_name` = ?, `last_name` = ? WHERE id = ?"
		assert.Equal(t, exp, q.query)
		assert.ElementsMatch(t, []string{"first_name", "last_name"}, q.fields)
		assert.ElementsMatch(t, []interface{}{"go", "land", 1}, q.values)
	})
	t.Run("tableが空の場合はエラー", func(t *testing.T) {
		q, err := BuildUpdateQuery("", nil, nil)
		assert.Nil(t, q)
		assert.Errorf(t, err, "empty table name")
	})
	t.Run("setが空の場合はエラー", func(t *testing.T) {
		q, err := BuildUpdateQuery("users", make(map[string]interface{}), nil)
		assert.Nil(t, q)
		assert.Errorf(t, err, "empty field set")
	})
	t.Run("Whereが空の場合はエラー", func(t *testing.T) {
		q, err := BuildUpdateQuery("users", map[string]interface{}{
			"first_name": "go",
		}, Where(""))
		assert.Nil(t, q)
		assert.Errorf(t, err, "empty where clause")
	})
}
