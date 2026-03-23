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
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.PrimaryUint64{}), false)

		assert.NoError(t, err)
		if !assert.NotNil(t, metadata) {
			return
		}
		if assert.NotNil(t, metadata.autoIncrementField) {
			assert.Equal(t, 0, *metadata.autoIncrementField)
		}
		assert.Equal(t, []int{0}, metadata.primaryKeyFields)
		assert.Equal(t, []int{1}, metadata.updatableFields)

		idIndex, ok := metadata.fields["id"]
		assert.True(t, ok)
		assert.Equal(t, 0, idIndex)

		nameIndex, ok := metadata.fields["name"]
		assert.True(t, ok)
		assert.Equal(t, 1, nameIndex)

		colName, ok := metadata.columns[1]
		assert.True(t, ok)
		assert.Equal(t, "name", colName)

		_, ok = metadata.fields["note"]
		assert.False(t, ok)
	})

	t.Run("supports multiple primary keys", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeOf(testmodel.MultiplePrimaryKey{}), false)

		assert.NoError(t, err)
		if !assert.NotNil(t, metadata) {
			return
		}
		assert.Nil(t, metadata.autoIncrementField)
		assert.Equal(t, []int{0, 1}, metadata.primaryKeyFields)
		assert.Equal(t, []int{0, 1, 2}, metadata.updatableFields)

		pk1Index, ok := metadata.fields["pk1"]
		assert.True(t, ok)
		assert.Equal(t, 0, pk1Index)

		pk2Name, ok := metadata.columns[1]
		assert.True(t, ok)
		assert.Equal(t, "pk2", pk2Name)
	})

	t.Run("returns error when no exql tags are defined", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeFor[testmodel.NoTag](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "no exql tags in any fields")
	})

	t.Run("returns error when no primary key is defined", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeFor[testmodel.NoPrimaryKey](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "table has no primary key")
	})

	t.Run("returns error when column tag is not set", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeFor[testmodel.NoColumnTag](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "column tag is not set")
	})

	t.Run("returns error for pointer fields", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeFor[testmodel.UpdateSample](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "field must not be a pointer:  ptr")
	})

	t.Run("returns error for invalid tag format", func(t *testing.T) {
		metadata, err := aggregateFields(reflect.TypeFor[testmodel.BadTag](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "duplicated tag: a")
	})
}

func TestAggregateModelValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		schema, _ := noCacheReflector.GetSchema(reflect.TypeFor[model.Users](), false)
		user := &model.Users{Name: "go", Age: 10}
		m, err := schema.aggregateModelValue(user)
		assert.NoError(t, err)
		assert.NotNil(t, m.autoIncrementField)
		assert.ElementsMatch(t, []string{"age", "name"}, m.values.Keys())
		assert.ElementsMatch(t, []any{int64(10), "go"}, m.values.Values())
	})
	t.Run("multiple primary key", func(t *testing.T) {
		schema, _ := noCacheReflector.GetSchema(reflect.TypeFor[testmodel.MultiplePrimaryKey](), false)
		data := &testmodel.MultiplePrimaryKey{
			Pk1:   "val1",
			Pk2:   "val2",
			Other: 1,
		}
		v, err := schema.aggregateModelValue(data)
		assert.NoError(t, err)
		assert.Nil(t, v.autoIncrementField)
		assert.ElementsMatch(t, []string{"pk1", "pk2", "other"}, v.values.Keys())
		assert.ElementsMatch(t, []any{"val1", "val2", 1}, v.values.Values())
	})
	assertInvalid := func(t *testing.T, m Model, e string) {
		s, f, err := QueryForInsert(m)
		assert.Nil(t, s)
		assert.Nil(t, f)
		assert.EqualError(t, err, e)
	}
	t.Run("should error if dest is nil", func(t *testing.T) {
		assertInvalid(t, nil, errModelNil.Error())
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
		assertInvalid(t, &testmodel.NoTag{}, "no exql tags in any fields")
	})
}

func TestAggregateModelUpdateValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		name := "go"
		age := int64(20)
		schema, err := aggregateFields(reflect.TypeFor[model.UpdateUsers](), true)
		assert.NoError(t, err)
		if !assert.NotNil(t, schema) {
			return
		}

		v, err := schema.aggregateModelUpdateValue(&model.UpdateUsers{
			Name: &name,
			Age:  &age,
		})
		assert.NoError(t, err)
		if !assert.NotNil(t, v) {
			return
		}
		assert.Equal(t, "users", v.tableName)
		assert.NotNil(t, v.autoIncrementField)
		assert.ElementsMatch(t, []string{"age", "name"}, v.values.Keys())
		assert.ElementsMatch(t, []any{int64(20), "go"}, v.values.Values())
	})

	t.Run("no non nil update fields", func(t *testing.T) {
		schema, err := aggregateFields(reflect.TypeFor[testmodel.UpdateSample](), true)
		assert.NoError(t, err)
		if !assert.NotNil(t, schema) {
			return
		}

		v, err := schema.aggregateModelUpdateValue(&testmodel.UpdateSample{})
		assert.Nil(t, v)
		assert.EqualError(t, err, "no updatable fields with non-nil value")
	})
}

func TestCreateReceivers(t *testing.T) {
	schema, err := aggregateFields(reflect.TypeFor[model.Users](), false)
	assert.NoError(t, err)
	if !assert.NotNil(t, schema) {
		return
	}

	dest := reflect.ValueOf(&model.Users{}).Elem()
	receivers := schema.createReceivers([]string{"id", "unknown", "name", "age"}, &dest)

	if assert.Len(t, receivers, 4) {
		idReceiver, ok := receivers[0].(*int64)
		if assert.True(t, ok) {
			assert.Same(t, dest.FieldByName("Id").Addr().Interface(), idReceiver)
			*idReceiver = 10
		}

		_, ok = receivers[1].(*noopScanner)
		assert.True(t, ok)

		nameReceiver, ok := receivers[2].(*string)
		if assert.True(t, ok) {
			assert.Same(t, dest.FieldByName("Name").Addr().Interface(), nameReceiver)
			*nameReceiver = "alice"
		}

		ageReceiver, ok := receivers[3].(*int64)
		if assert.True(t, ok) {
			assert.Same(t, dest.FieldByName("Age").Addr().Interface(), ageReceiver)
			*ageReceiver = 20
		}
	}

	assert.Equal(t, model.Users{
		Id:   10,
		Name: "alice",
		Age:  20,
	}, dest.Interface())
}
