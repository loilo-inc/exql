package exql

import (
	"reflect"
	"sync"
	"testing"

	"github.com/loilo-inc/exql/v3/model"
	"github.com/stretchr/testify/assert"
)

func Test_resolveDestination(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, err := resolveDestination(nil)
		assert.ErrorIs(t, errMapDestination, err)
	})
	t.Run("non-pointer", func(t *testing.T) {
		_, err := resolveDestination(1)
		assert.ErrorIs(t, errMapDestination, err)
	})
	t.Run("pointer of non-struct", func(t *testing.T) {
		var i int
		_, err := resolveDestination(&i)
		assert.ErrorIs(t, errMapDestination, err)
	})
	t.Run("*struct", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		v, err := resolveDestination(u)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), v.FieldByName("Id").Int())
		assert.Equal(t, "alice", v.FieldByName("Name").String())
	})
	t.Run("**struct", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		_, err := resolveDestination(&u)
		assert.ErrorIs(t, errMapDestination, err)
	})
}

func Test_resolveNullableDestination(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, err := resolveNullableDestination(nil)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
	t.Run("non-pointer", func(t *testing.T) {
		_, err := resolveNullableDestination(1)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
	t.Run("pointer of non-struct", func(t *testing.T) {
		var i int
		_, err := resolveNullableDestination(&i)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
	t.Run("*struct", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		v, err := resolveNullableDestination(u)
		assert.NoError(t, err)
		assert.Equal(t, reflect.Struct, v.elemType.Kind())
		assert.Equal(t, reflect.Struct, v.value.Kind())
		assert.Equal(t, int64(1), v.value.FieldByName("Id").Int())
		assert.Equal(t, "alice", v.value.FieldByName("Name").String())
	})
	t.Run("**struct(nil)", func(t *testing.T) {
		var u *model.Users
		v, err := resolveNullableDestination(&u)
		assert.NoError(t, err)
		assert.Equal(t, reflect.Struct, v.elemType.Kind())
		assert.Equal(t, reflect.Pointer, v.value.Kind())
		assert.True(t, v.value.IsNil())
	})
	t.Run("**struct(non-nil)", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		_, err := resolveNullableDestination(&u)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
}

func Test_resolveDestinationMany(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, _, err := resolveDestinationMany(nil)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("non-pointer", func(t *testing.T) {
		_, _, err := resolveDestinationMany([]*model.Users{})
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of non-slice", func(t *testing.T) {
		var u model.Users
		_, _, err := resolveDestinationMany(&u)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of slice of non-pointer", func(t *testing.T) {
		var users []model.Users
		_, _, err := resolveDestinationMany(&users)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of slice of pointer of non-struct", func(t *testing.T) {
		var values []*int
		_, _, err := resolveDestinationMany(&values)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of slice of pointer of struct", func(t *testing.T) {
		users := []*model.Users{{Id: 1, Name: "alice"}}
		typ, v, err := resolveDestinationMany(&users)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeFor[model.Users](), typ)
		assert.Equal(t, reflect.Pointer, v.Kind())
		assert.Equal(t, reflect.Slice, v.Elem().Kind())
		assert.Len(t, v.Elem().Interface(), 1)
		assert.Equal(t, "alice", v.Elem().Index(0).Elem().FieldByName("Name").String())
	})
}

func TestReflector_getSchema(t *testing.T) {
	t.Run("returns schema for model pointer", func(t *testing.T) {
		r := newReflector()

		schema, err := r.getModelSchema(&model.Users{}, false)

		assert.NoError(t, err)
		if !assert.NotNil(t, schema) {
			return
		}
		if assert.NotNil(t, schema.autoIncrementField) {
			assert.Equal(t, 0, *schema.autoIncrementField)
		}
		assert.Equal(t, []int{1, 2}, schema.updatableFields)
	})

	t.Run("returns schema for model update struct", func(t *testing.T) {
		r := newReflector()

		schema, err := r.getModelSchema(&model.UpdateUsers{}, true)

		assert.NoError(t, err)
		if !assert.NotNil(t, schema) {
			return
		}
		if assert.NotNil(t, schema.autoIncrementField) {
			assert.Equal(t, 0, *schema.autoIncrementField)
		}
		assert.Equal(t, []int{1, 2}, schema.updatableFields)
	})

	t.Run("returns validation error for invalid destination", func(t *testing.T) {
		r := newReflector()

		schema, err := r.getModelSchema(model.UpdateUsers{}, false)

		assert.Nil(t, schema)
		assert.ErrorIs(t, err, errMapDestination)
	})

	t.Run("uses cached schema when cache is enabled", func(t *testing.T) {
		r := newReflector()

		s1, err := r.getModelSchema(&model.Users{}, false)
		assert.NoError(t, err)

		s2, err := r.getModelSchema(&model.Users{}, false)
		assert.NoError(t, err)
		assert.Same(t, s1, s2)
	})

	t.Run("rebuilds schema when cache is disabled", func(t *testing.T) {
		r := &reflector{cache: false}

		s1, err := r.getModelSchema(&model.Users{}, false)
		assert.NoError(t, err)

		s2, err := r.getModelSchema(&model.Users{}, false)
		assert.NoError(t, err)
		assert.NotSame(t, s1, s2)
	})

	t.Run("is safe for concurrent access", func(t *testing.T) {
		r := newReflector()
		const goroutines = 64
		expected, err := r.getModelSchema(&model.Users{}, false)
		assert.NoError(t, err)

		results := make(chan *modelSchema, goroutines)
		errs := make(chan error, goroutines)
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for range goroutines {
			go func() {
				defer wg.Done()
				schema, err := r.getModelSchema(&model.Users{}, false)
				if err != nil {
					errs <- err
					return
				}
				results <- schema
			}()
		}

		wg.Wait()
		close(results)
		close(errs)

		for err := range errs {
			assert.NoError(t, err)
		}

		var first *modelSchema
		count := 0
		for schema := range results {
			if !assert.NotNil(t, schema) {
				continue
			}
			assert.Same(t, expected, schema)
			first = schema
			count++
		}
		assert.Same(t, expected, first)
		assert.Equal(t, goroutines, count)
	})
}

func TestReflectorclearSchemaCache(t *testing.T) {
	r := newReflector()

	s1, err := r.getSchema(reflect.TypeFor[model.Users](), false)
	assert.NoError(t, err)

	r.clearSchemaCache()

	s2, err := r.getSchema(reflect.TypeFor[model.Users](), false)
	assert.NoError(t, err)
	assert.NotSame(t, s1, s2)
}

func Test_typeKey(t *testing.T) {
	assert.Equal(t, "github.com/loilo-inc/exql/v3/model.Users", typeKey(reflect.TypeFor[model.Users]()))
}
