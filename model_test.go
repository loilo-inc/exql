package exql

import (
	"reflect"
	"testing"

	"github.com/loilo-inc/exql/v3/model"
	"github.com/loilo-inc/exql/v3/model/testmodel"
	"github.com/stretchr/testify/assert"
)

func TestAggregateUpsertSchema(t *testing.T) {
	t.Run("builds model metadata from exql tags", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[model.Users](), false)

		assert.NoError(t, err)
		if !assert.NotNil(t, metadata) {
			return
		}
		if assert.NotNil(t, metadata.autoIncrementField) {
			assert.Equal(t, 0, *metadata.autoIncrementField)
		}
		assert.Equal(t, []column{
			{index: 1, name: "name"},
			{index: 2, name: "age"},
		}, metadata.columns)
	})

	t.Run("supports multiple primary keys", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[testmodel.MultiplePrimaryKey](), false)

		assert.NoError(t, err)
		if !assert.NotNil(t, metadata) {
			return
		}
		assert.Nil(t, metadata.autoIncrementField)
		assert.Equal(t, []column{
			{index: 0, name: "pk1"},
			{index: 1, name: "pk2"},
			{index: 2, name: "other"},
		}, metadata.columns)
	})

	t.Run("returns error when type is not struct", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[int](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "type must be struct")
	})

	t.Run("returns error when no exql tags are defined", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[testmodel.NoTag](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "no exql tags in any fields")
	})

	t.Run("returns error when column tag is not set", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[testmodel.NoColumnTag](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "column tag is not set")
	})

	t.Run("returns error for pointer fields", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[testmodel.UpdateSample](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "field must not be a pointer:  ptr")
	})

	t.Run("returns error for invalid tag format", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[testmodel.BadTag](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "duplicated tag: a")
	})

	t.Run("returns error for auto_increment field with non int64 type", func(t *testing.T) {
		metadata, err := parseUpsertSchema(reflect.TypeFor[testmodel.PrimaryUint64](), false)

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "auto_increment field must be int64")
	})
}

func Test_UpsertSchema_aggregateValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		schema, _ := parseUpsertSchema(reflect.TypeFor[model.Users](), false)
		user := &model.Users{Name: "go", Age: 10}
		m, err := schema.aggregateValue(user)
		assert.NoError(t, err)
		assert.NotNil(t, m.autoIncrementField)
		assert.Equal(t, map[string]any{
			"name": "go",
			"age":  int64(10),
		}, m.values)
	})
	t.Run("multiple primary key", func(t *testing.T) {
		schema, _ := parseUpsertSchema(reflect.TypeFor[testmodel.MultiplePrimaryKey](), false)
		data := &testmodel.MultiplePrimaryKey{
			Pk1:   "val1",
			Pk2:   "val2",
			Other: 1,
		}
		v, err := schema.aggregateValue(data)
		assert.NoError(t, err)
		assert.Nil(t, v.autoIncrementField)
		assert.Equal(t, map[string]any{
			"pk1":   "val1",
			"pk2":   "val2",
			"other": 1,
		}, v.values)
	})

	t.Run("should error if dest is nil", func(t *testing.T) {
		schema, _ := parseUpsertSchema(reflect.TypeFor[model.Users](), false)
		_, err := schema.aggregateValue(nil)
		assert.EqualError(t, err, errModelNil.Error())
	})
}

func Test_AggregateMapSchema(t *testing.T) {
	t.Run("builds model metadata from exql tags", func(t *testing.T) {
		metadata, err := parseMapSchema(reflect.TypeFor[model.Users]())

		assert.NoError(t, err)
		if !assert.NotNil(t, metadata) {
			return
		}
		idIndex, ok := metadata.fields["id"]
		assert.True(t, ok)
		assert.Equal(t, 0, idIndex)

		nameIndex, ok := metadata.fields["name"]
		assert.True(t, ok)
		assert.Equal(t, 1, nameIndex)

		_, ok = metadata.fields["note"]
		assert.False(t, ok)
	})
	t.Run("returns error when type is not struct", func(t *testing.T) {
		metadata, err := parseMapSchema(reflect.TypeFor[int]())

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "type must be struct")
	})
	t.Run("returns error when no exql tags are defined", func(t *testing.T) {
		metadata, err := parseMapSchema(reflect.TypeFor[testmodel.NoTag]())

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "no exql tags in any fields")
	})
	t.Run("returns error when column tag is not set", func(t *testing.T) {
		metadata, err := parseMapSchema(reflect.TypeFor[testmodel.NoColumnTag]())

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "column tag is not set")
	})
	t.Run("returns error for invalid tag format", func(t *testing.T) {
		metadata, err := parseMapSchema(reflect.TypeFor[testmodel.BadTag]())

		assert.Nil(t, metadata)
		assert.EqualError(t, err, "duplicated tag: a")
	})
}

func Test_MapSchema_CreateReceivers(t *testing.T) {
	schema, err := parseMapSchema(reflect.TypeFor[model.Users]())
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
