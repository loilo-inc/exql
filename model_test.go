package exql

import (
	"reflect"
	"testing"

	"github.com/loilo-inc/exql/v3/model"
	"github.com/loilo-inc/exql/v3/model/testmodel"
	"github.com/stretchr/testify/assert"
)

func TestAggregateFields(t *testing.T) {
	t.Run("builds model metadata from exql tags", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.PrimaryUint64{}))

		assert.NoError(t, err)
		if !assert.NotNil(t, metadata) {
			return
		}
		if assert.NotNil(t, metadata.autoIncrementField) {
			assert.Equal(t, 0, *metadata.autoIncrementField)
		}
		assert.Equal(t, []int{0}, metadata.primaryKeyFields)
		assert.Equal(t, []int{1}, metadata.updatableFields)

		idIndex, ok := metadata.fields.Load("id")
		assert.True(t, ok)
		assert.Equal(t, 0, idIndex)

		nameIndex, ok := metadata.fields.Load("name")
		assert.True(t, ok)
		assert.Equal(t, 1, nameIndex)

		colName, ok := metadata.columns.Load(1)
		assert.True(t, ok)
		assert.Equal(t, "name", colName)

		_, ok = metadata.fields.Load("note")
		assert.False(t, ok)
	})

	t.Run("supports multiple primary keys", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.MultiplePrimaryKey{}))

		assert.NoError(t, err)
		if !assert.NotNil(t, metadata) {
			return
		}
		assert.Nil(t, metadata.autoIncrementField)
		assert.Equal(t, []int{0, 1}, metadata.primaryKeyFields)
		assert.Equal(t, []int{0, 1, 2}, metadata.updatableFields)

		pk1Index, ok := metadata.fields.Load("pk1")
		assert.True(t, ok)
		assert.Equal(t, 0, pk1Index)

		pk2Name, ok := metadata.columns.Load(1)
		assert.True(t, ok)
		assert.Equal(t, "pk2", pk2Name)
	})

	t.Run("returns error when no exql tags are defined", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.NoTag{}))

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "obj doesn't have exql tags in any fields")
	})

	t.Run("returns error when no primary key is defined", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.NoPrimaryKey{}))

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "table has no primary key")
	})

	t.Run("returns error when column tag is not set", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.NoColumnTag{}))

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "column tag is not set")
	})

	t.Run("returns error for pointer fields", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.UpdateSample{}))

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "struct field must not be a pointer:  ptr")
	})

	t.Run("returns error for invalid tag format", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.BadTag{}))

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "duplicated tag: a")
	})
}

func TestAggregateModelValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		user := &model.Users{Name: "go", Age: 10}
		schema, _ := defaultReflector().GetSchema(user)
		m, err := schema.aggregateModelValue(user)
		assert.NoError(t, err)
		assert.NotNil(t, m.autoIncrementField)
		assert.ElementsMatch(t, []string{"age", "name"}, m.values.Keys())
		assert.ElementsMatch(t, []any{int64(10), "go"}, m.values.Values())
	})
	t.Run("multiple primary key", func(t *testing.T) {
		data := &testmodel.MultiplePrimaryKey{
			Pk1:   "val1",
			Pk2:   "val2",
			Other: 1,
		}
		schema, _ := defaultReflector().GetSchema(data)
		v, err := schema.aggregateModelValue(data)
		assert.NoError(t, err)
		assert.Nil(t, v.autoIncrementField)
		assert.ElementsMatch(t, []string{"pk1", "pk2", "other"}, v.values.Keys())
		assert.ElementsMatch(t, []any{"val1", "val2", 1}, v.values.Values())
	})
	assertInvalid := func(t *testing.T, m Model, e string) {
		s, f, err := QueryForInsert(defaultReflector(), m)
		assert.Nil(t, s)
		assert.Nil(t, f)
		assert.EqualError(t, err, e)
	}
	t.Run("should error if dest is nil", func(t *testing.T) {
		assertInvalid(t, nil, "model is nil")
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
