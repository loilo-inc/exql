package exql_test

import (
	"testing"

	"github.com/loilo-inc/exql/v2"
	"github.com/loilo-inc/exql/v2/model"
	"github.com/loilo-inc/exql/v2/model/testmodel"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null"
)

func TestQueryWhere(t *testing.T) {
	t.Run("Where", func(t *testing.T) {
		v, args, err := exql.Where("q = ?", 1).Query()
		assert.NoError(t, err)
		assert.Equal(t, "q = ?", v)
		assert.ElementsMatch(t, []any{1}, args)
	})
}
func TestQueryForInsert(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		user := model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("name"),
		}
		s, f, err := exql.QueryForInsert(&user)
		assert.NoError(t, err)
		assert.NotNil(t, f)
		exp := "INSERT INTO `users` (`first_name`,`last_name`) VALUES (?,?)"
		stmt, args, err := s.Query()
		assert.NoError(t, err)
		assert.Equal(t, exp, stmt)
		assert.ElementsMatch(t, args, []any{user.FirstName, user.LastName})
	})
}

func TestQueryForBulkInsert(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		q, err := exql.QueryForBulkInsert(
			&model.Users{FirstName: null.StringFrom("user"), LastName: null.StringFrom("one")},
			&model.Users{FirstName: null.StringFrom("user"), LastName: null.StringFrom("two")},
		)
		assert.NoError(t, err)
		stmt, args, err := q.Query()
		assert.NoError(t, err)
		assert.Equal(t, "INSERT INTO `users` (`first_name`,`last_name`) VALUES (?,?),(?,?)", stmt)
		assert.ElementsMatch(t, []any{
			null.StringFrom("user"), null.StringFrom("one"),
			null.StringFrom("user"), null.StringFrom("two"),
		}, args)
	})
	t.Run("error if args empty", func(t *testing.T) {
		q, err := exql.QueryForBulkInsert[model.Users]()
		assert.Nil(t, q)
		assert.EqualError(t, err, "empty list")
	})
}

func TestAggregateModelMetadata(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		m, err := exql.AggregateModelMetadata(&model.Users{
			FirstName: null.StringFrom("first"),
			LastName:  null.StringFrom("name"),
		})
		assert.NoError(t, err)
		assert.Equal(t, "users", m.TableName)
		assert.NotNil(t, m.AutoIncrementField)
		assert.ElementsMatch(t, []string{"id"}, m.PrimaryKeyColumns)
		assert.ElementsMatch(t, []any{int64(0)}, m.PrimaryKeyValues)
		assert.ElementsMatch(t, []string{"first_name", "last_name"}, m.Values.Keys())
		assert.ElementsMatch(t, []any{null.StringFrom("first"), null.StringFrom("name")}, m.Values.Values())
	})
	t.Run("multiple primary key", func(t *testing.T) {
		data := &testmodel.MultiplePrimaryKey{
			Pk1:   "val1",
			Pk2:   "val2",
			Other: 1,
		}
		md, err := exql.AggregateModelMetadata(data)
		assert.NoError(t, err)
		assert.Equal(t, data.TableName(), md.TableName)
		assert.Nil(t, md.AutoIncrementField)
		assert.ElementsMatch(t, []string{"pk1", "pk2"}, md.PrimaryKeyColumns)
		assert.ElementsMatch(t, []any{"val1", "val2"}, md.PrimaryKeyValues)
		assert.ElementsMatch(t, []string{"pk1", "pk2", "other"}, md.Values.Keys())
		assert.ElementsMatch(t, []any{"val1", "val2", 1}, md.Values.Values())
	})
	assertInvalid := func(t *testing.T, m exql.Model, e string) {
		s, f, err := exql.QueryForInsert(m)
		assert.Nil(t, s)
		assert.Nil(t, f)
		assert.EqualError(t, err, e)
	}
	t.Run("should error if dest is nil", func(t *testing.T) {
		assertInvalid(t, nil, "pointer is nil")
	})
	t.Run("should error if dest is not pointer", func(t *testing.T) {
		user := model.Users{}
		assertInvalid(t, user, "object must be pointer of struct")
	})
	t.Run("should error if TableName() doesn't return string", func(t *testing.T) {
		assertInvalid(t, &testmodel.BadTableName{}, "empty table name")
	})
	t.Run("should error if field doesn't have column tag", func(t *testing.T) {
		assertInvalid(t, &testmodel.NoColumnTag{}, "column tag is not set")
	})
	t.Run("should error if field tag is invalid", func(t *testing.T) {
		assertInvalid(t, &testmodel.BadTag{}, "duplicated tag: a")
	})
	t.Run("should error if dest has no primary key tag", func(t *testing.T) {
		assertInvalid(t, &testmodel.NoPrimaryKey{}, "table has no primary key")
	})
	t.Run("shoud error if no exql tags found", func(t *testing.T) {
		assertInvalid(t, &testmodel.NoTag{}, "obj doesn't have exql tags in any fields")
	})
}

func TestQueryForUpdateModel(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		user := &model.Users{}
		user.FirstName.SetValid("new")
		user.LastName.SetValid("name")
		q, err := exql.QueryForUpdateModel(&model.UpdateUsers{
			FirstName: &user.FirstName,
			LastName:  &user.LastName,
		}, exql.Where(`id = ?`, 1))
		if err != nil {
			t.Fatal(err)
		}
		stmt, args, err := q.Query()
		assert.NoError(t, err)
		assert.Equal(t, stmt,
			"UPDATE `users` SET `first_name` = ?,`last_name` = ? WHERE id = ?",
		)
		assert.ElementsMatch(t, []any{user.FirstName, user.LastName, 1}, args)
	})
	t.Run("should error if pointer is nil", func(t *testing.T) {
		_, err := exql.QueryForUpdateModel(nil, nil)
		assert.EqualError(t, err, "pointer is nil")
	})
	t.Run("should error if not pointer", func(t *testing.T) {
		_, err := exql.QueryForUpdateModel(model.UpdateUsers{}, nil)
		assert.EqualError(t, err, "must be pointer of struct")
	})
	t.Run("should error if has invalid tag", func(t *testing.T) {
		_, err := exql.QueryForUpdateModel(&upSampleInvalidTag{}, nil)
		assert.EqualError(t, err, "invalid tag format")
	})
	t.Run("should error if field is not pointer", func(t *testing.T) {
		_, err := exql.QueryForUpdateModel(&upSampleNotPtr{}, nil)
		assert.EqualError(t, err, "field must be pointer")
	})
	t.Run("should ignore if field is nil", func(t *testing.T) {
		_, err := exql.QueryForUpdateModel(&upSample{}, nil)
		assert.EqualError(t, err, "no value for update")
	})
	t.Run("should error if struct has no fields", func(t *testing.T) {
		_, err := exql.QueryForUpdateModel(&upSampleNoFields{}, nil)
		assert.EqualError(t, err, "struct has no field")
	})
	t.Run("should error if struct doesn't implement ForTableName()", func(t *testing.T) {
		id := 1
		_, err := exql.QueryForUpdateModel(&upSample{Id: &id}, nil)
		assert.EqualError(t, err, "empty table name")
	})
	t.Run("should error if no column in tag", func(t *testing.T) {
		id := 1
		_, err := exql.QueryForUpdateModel(&upSampleNoColumn{Id: &id}, nil)
		assert.EqualError(t, err, "tag must include column")
	})
}
