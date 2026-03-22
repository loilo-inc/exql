package exql

import (
	"testing"

	"github.com/loilo-inc/exql/v3/model"
	"github.com/loilo-inc/exql/v3/model/testmodel"
	"github.com/stretchr/testify/assert"
)

func TestQueryWhere(t *testing.T) {
	t.Run("Where", func(t *testing.T) {
		v, args, err := Where("q = ?", 1).Query()
		assert.NoError(t, err)
		assert.Equal(t, "q = ?", v)
		assert.ElementsMatch(t, []any{1}, args)
	})
}

func TestQueryForInsert(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		user := model.Users{
			Name: "go", Age: 10,
		}
		s, f, err := QueryForInsert(&user)
		assert.NoError(t, err)
		assert.NotNil(t, f)
		exp := "INSERT INTO `users` (`age`,`name`) VALUES (?,?)"
		stmt, args, err := s.Query()
		assert.NoError(t, err)
		assert.Equal(t, exp, stmt)
		assert.ElementsMatch(t, args, []any{user.Age, user.Name})
	})
	t.Run("should error if Reflector returns error", func(t *testing.T) {
		s, f, err := QueryForInsert(&testmodel.NoTag{})
		assert.Nil(t, s)
		assert.Nil(t, f)
		assert.EqualError(t, err, "obj doesn't have exql tags in any fields")
	})
	t.Run("should error if injected Reflector returns error", func(t *testing.T) {
		s, f, err := queryForInsert(&errReflector{}, &model.Users{})
		assert.Nil(t, s)
		assert.Nil(t, f)
		assert.EqualError(t, err, "error reflector")
	})
}

func TestQueryForBulkInsert(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		q, err := QueryForBulkInsert(
			&model.Users{Age: 1, Name: "one"},
			&model.Users{Age: 2, Name: "two"},
		)
		assert.NoError(t, err)
		stmt, args, err := q.Query()
		assert.NoError(t, err)
		assert.Equal(t, "INSERT INTO `users` (`age`,`name`) VALUES (?,?),(?,?)", stmt)
		assert.ElementsMatch(t, []any{int64(1), "one", int64(2), "two"}, args)
	})
	t.Run("error if args empty", func(t *testing.T) {
		q, err := QueryForBulkInsert[*model.Users]()
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty list")
	})
	t.Run("should error if injected Reflector returns error", func(t *testing.T) {
		q, err := queryForBulkInsert(&errReflector{}, &model.Users{})
		assert.Nil(t, q)
		assert.EqualError(t, err, "error reflector")
	})
}

func TestQueryForUpdateModel(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		name := "go"
		age := int64(20)
		q, err := QueryForUpdateModel(&model.UpdateUsers{
			Name: &name,
			Age:  &age,
		}, Where(`id = ?`, 1))
		if err != nil {
			t.Fatal(err)
		}
		stmt, args, err := q.Query()
		assert.NoError(t, err)
		assert.Equal(t, stmt,
			"UPDATE `users` SET `age` = ?,`name` = ? WHERE id = ?",
		)
		assert.ElementsMatch(t, []any{age, name, 1}, args)
	})
	t.Run("should error if model is nil", func(t *testing.T) {
		_, err := QueryForUpdateModel(nil, nil)
		assert.EqualError(t, err, "model is nil")
	})
	t.Run("should error if has invalid tag", func(t *testing.T) {
		_, err := QueryForUpdateModel(&testmodel.UpdateSampleInvalidTag{}, nil)
		assert.EqualError(t, err, "invalid tag format")
	})
	t.Run("should error if field is not pointer", func(t *testing.T) {
		_, err := QueryForUpdateModel(&testmodel.UpdateSampleNotPtr{}, nil)
		assert.EqualError(t, err, "field must be pointer")
	})
	t.Run("should ignore if field is nil", func(t *testing.T) {
		_, err := QueryForUpdateModel(&testmodel.UpdateSample{}, nil)
		assert.EqualError(t, err, "no value for update")
	})
	t.Run("should error if struct has no fields", func(t *testing.T) {
		_, err := QueryForUpdateModel(&testmodel.UpdateSampleNoFields{}, nil)
		assert.EqualError(t, err, "struct has no field")
	})
	t.Run("should error if struct doesn't implement ForTableName()", func(t *testing.T) {
		id := 1
		_, err := QueryForUpdateModel(&testmodel.UpdateSample{Id: &id}, nil)
		assert.EqualError(t, err, "empty table name")
	})
	t.Run("should error if no column in tag", func(t *testing.T) {
		id := 1
		_, err := QueryForUpdateModel(&testmodel.UpdateSampleNoColumn{Id: &id}, nil)
		assert.EqualError(t, err, "tag must include column")
	})
}
